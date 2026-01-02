package uartsdlc

import (
	"encoding/hex"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
	"rvpro3/radarvision.com/utils"
)

var errWriteMessageDiscarded = errors.New("write message discarded")

const SDLCServiceMetricsAt = "SDLC.UART"
const SDLCServiceName = "SDLC.Service"

const sdlcUARTEnabled = "UART.UART.Enabled"
const sdlcUARTPortName = "SDLC.UART.PortName"
const sdlcUARTBaudRate = "SDLC.UART.BaudRate"
const sdlcUARTDataBits = "SDLC.UART.DataBits"
const sdlcUARTParity = "SDLC.UART.Parity"
const sdlcUARTStopBits = "SDLC.UART.StopBits"
const sdlcUARTCSVEnabled = "SDLC.UART.CSV.Enabled"
const sdlcUARTCSVFilePathTemplate = "SDLC.UART.CSVFilePathTemplate"
const sdlcUARTCSVFilePathFormat = "SDLC.UART.CSVFilePathFormat"
const writeAction = "w"
const readAction = "r"
const errorAction = "e"

type SDLCService struct {
	Serial                    SerialConnection
	doneChan                  chan bool
	writeChannel              chan []byte
	readBuffer                [1024]byte
	backingBuffer             [2048]byte
	serialBuffer              utils.SerialBuffer
	terminate                 bool
	terminateRefCount         atomic.Int32
	WritePool                 *SDLCWritePool             `json:"-"`
	OnError                   func(*SDLCService, error)  `json:"-"`
	OnTerminate               func(*SDLCService)         `json:"-"`
	OnReadMessage             func(*SDLCService, []byte) `json:"-"`
	OnWriteMessage            func(*SDLCService, []byte) `json:"-"`
	IsWriteEnabledMetric      *utils.Metric
	WriteEnqueuedMetric       *utils.Metric
	WriteEnqueuedBytesMetric  *utils.Metric
	WriteDequeuedMetric       *utils.Metric
	WriteDequeuedBytesMetric  *utils.Metric
	WriteQueueFullMetric      *utils.Metric
	WriteQueueFullBytesMetric *utils.Metric
	WriteSuccessMetric        *utils.Metric
	WriteSuccessBytesMetric   *utils.Metric
	WriteErrorMetric          *utils.Metric
	WriteErrorBytesMetric     *utils.Metric
	WriteOmitMetric           *utils.Metric
	WriteOmitBytesMetric      *utils.Metric
	OmitLogWritesMetric       *utils.Metric
	OmitLogReadsMetric        *utils.Metric
	RetrySleepDuration        time.Duration
	Error                     error
	CsvProvider               *utils.CSVRollOverFileWriterProvider `json:"-"`
	ReadBytesMetric           *utils.Metric
	ReadsMetric               *utils.Metric
	PopsMetric                *utils.Metric
	PopBytesMetric            *utils.Metric
}

func (s *SDLCService) SetupDefaults(config *utils.Config) {
	s.init()
	config.SetSettingAsBool(sdlcUARTEnabled, true)
	config.SetSettingAsStr(sdlcUARTPortName, "/dev/ttymxc2")
	config.SetSettingAsInt(sdlcUARTBaudRate, 115200)
	config.SetSettingAsInt(sdlcUARTDataBits, 8)
	config.SetSettingAsInt(sdlcUARTParity, 0)
	config.SetSettingAsInt(sdlcUARTStopBits, 0)
	config.SetSettingAsBool(sdlcUARTCSVEnabled, true)
	config.SetSettingAsStr(sdlcUARTCSVFilePathTemplate, "/media/SDLOGS/logs/system/uart-%s.csv")
	config.SetSettingAsStr(sdlcUARTCSVFilePathFormat, "20060102")
}

func (s *SDLCService) SetupRunnable(state *utils.State, config *utils.Config) {
	if !config.GetSettingAsBool(sdlcUARTEnabled) {
		log.Info().Msg("SDLC UART is disabled")
		return
	}

	s.InitFromConfig(config)
	s.Start()

	state.Set(SDLCServiceName, s)
}

func (s *SDLCService) InitFromConfig(config *utils.Config) {
	s.Serial.PortName = config.GetSettingAsStr(sdlcUARTPortName)
	s.Serial.Mode.BaudRate = config.GetSettingAsInt(sdlcUARTBaudRate)
	s.Serial.Mode.DataBits = config.GetSettingAsInt(sdlcUARTDataBits)
	s.Serial.Mode.Parity = serial.Parity(config.GetSettingAsInt(sdlcUARTParity))
	s.Serial.Mode.StopBits = serial.StopBits(config.GetSettingAsInt(sdlcUARTStopBits))

	if config.GetSettingAsBool(sdlcUARTCSVEnabled) {
		s.CsvProvider = utils.NewCSVRollOverFileWriterProvider(
			config.GetSettingAsStr(sdlcUARTCSVFilePathTemplate),
			config.GetSettingAsStr(sdlcUARTCSVFilePathFormat),
			s.WriteLogHeader,
		)
	}
}

func (s *SDLCService) GetServiceName() string {
	return SDLCServiceName
}

func (s *SDLCService) GetServiceNames() []string {
	return nil
}

func (s *SDLCService) init() {
	s.RetrySleepDuration = time.Duration(1) * time.Second
	s.WritePool = NewSDLCWritePool()
	s.Serial.RetryGuard.RetryEvery = 3
	s.serialBuffer.Buffer = s.backingBuffer[:]
	s.serialBuffer.StartDelim = 0x02
	s.serialBuffer.EndDelim = 0x03
	s.doneChan = make(chan bool)
	s.writeChannel = make(chan []byte, 5)
	s.terminateRefCount.Store(2)
	s.initMetrics()
}

func (s *SDLCService) initMetrics() {
	gm := &utils.GlobalMetrics
	s.ReadBytesMetric = gm.U64(SDLCServiceMetricsAt, "Read Bytes")
	s.ReadsMetric = gm.U64(SDLCServiceMetricsAt, "Reads")
	s.PopsMetric = gm.U64(SDLCServiceMetricsAt, "Pops")
	s.PopBytesMetric = gm.U64(SDLCServiceMetricsAt, "Pop Bytes")

	s.WriteEnqueuedMetric = gm.U64(SDLCServiceMetricsAt, "Enqueued Writes")
	s.WriteEnqueuedBytesMetric = gm.U64(SDLCServiceMetricsAt, "Enqueued Write Bytes")
	s.WriteDequeuedMetric = gm.U64(SDLCServiceMetricsAt, "Dequeued Writes")
	s.WriteDequeuedBytesMetric = gm.U64(SDLCServiceMetricsAt, "Dequeued Write Bytes")
	s.WriteQueueFullMetric = gm.U64(SDLCServiceMetricsAt, "Error: Queue Full Writes rejected")
	s.WriteQueueFullBytesMetric = gm.U64(SDLCServiceMetricsAt, "Error: Queue Full Write Bytes rejected")
	s.WriteSuccessMetric = gm.U64(SDLCServiceMetricsAt, "Success Writes")
	s.WriteSuccessBytesMetric = gm.U64(SDLCServiceMetricsAt, "Success Write Bytes")
	s.WriteErrorMetric = gm.U64(SDLCServiceMetricsAt, "Error: Writes")
	s.WriteErrorBytesMetric = gm.U64(SDLCServiceMetricsAt, "Error: Write Bytes")
	s.WriteOmitMetric = gm.U64(SDLCServiceMetricsAt, "Omit Writes")
	s.WriteOmitBytesMetric = gm.U64(SDLCServiceMetricsAt, "Omit Write Bytes")
	s.OmitLogWritesMetric = gm.U64(SDLCServiceMetricsAt, "Omit Log Writes")
	s.OmitLogReadsMetric = gm.U64(SDLCServiceMetricsAt, "Omit Log Reads")
	s.IsWriteEnabledMetric = gm.U32(SDLCServiceMetricsAt, "IsWriteEnabled")
}

func (s *SDLCService) Start() {
	s.init()
	go s.executeReader()
	go s.executeWriter()
}

func (s *SDLCService) Stop() {
	s.doneChan <- true

	for s.terminateRefCount.Load() > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *SDLCService) executeReader() {
	s.IsWriteEnabledMetric.SetU32(1, time.Now())

	for !s.terminate {

		if s.Serial.Connect() {
			readSize := s.Serial.Read(s.readBuffer[:])
			if readSize > 0 {
				now := time.Now()
				s.ReadsMetric.Add(1, now)
				s.ReadBytesMetric.Add(readSize, now)

				if err := s.serialBuffer.Push(s.readBuffer[:readSize]); err != nil {
					s.logError(err)
					if s.OnError != nil {
						s.OnError(s, err)
					}
				} else {
					if readBytes := s.serialBuffer.Pop(); readBytes != nil {
						s.PopsMetric.Add(1, now)
						s.PopBytesMetric.Add(len(readBytes), now)

						s.logMessage(readBytes, readAction, s.OmitLogReadsMetric)
						if s.OnReadMessage != nil {
							s.OnReadMessage(s, readBytes)
						}
					}
				}
			} else {
				// No data was read.
				// Simply ignore
			}
		} else {
			time.Sleep(s.RetrySleepDuration)
		}
	}
	s.Serial.Disconnect()
	s.terminateRefCount.Add(-1)

	if s.OnTerminate != nil {
		s.OnTerminate(s)
	}
}

func (s *SDLCService) executeWriter() {
	for {
		select {
		case data := <-s.writeChannel:
			s.writeData(data)

		case <-s.doneChan:
			s.IsWriteEnabledMetric.SetU32(0, time.Now())
			s.terminate = true
			s.terminateRefCount.Add(-1)
			close(s.writeChannel)
			close(s.doneChan)
			return
		}
	}
}

func (s *SDLCService) writeData(data []byte) {
	now := time.Now()

	if s.IsWriteEnabledMetric.GetU32() == 1 {
		if s.Serial.Write(data) {
			s.WriteSuccessMetric.AddCount(1, now)
			s.WriteSuccessBytesMetric.AddCount(uint64(len(data)), now)
			s.logMessage(data, writeAction, s.OmitLogWritesMetric)
		} else {
			s.WriteErrorMetric.AddCount(1, now)
			s.WriteErrorBytesMetric.AddCount(uint64(len(data)), now)
		}
	} else {
		s.WriteOmitMetric.AddCount(1, now)
		s.WriteOmitBytesMetric.AddCount(uint64(len(data)), now)
	}

	s.WritePool.Release(data)
	s.WriteDequeuedBytesMetric.AddCount(uint64(len(data)), now)
	s.WriteDequeuedMetric.AddCount(1, now)
}

func (s *SDLCService) Write(data []byte) {
	if !s.terminate {
		now := time.Now()

		if len(s.writeChannel) < cap(s.writeChannel) {
			buffer := s.WritePool.Alloc()
			copy(buffer[0:len(data)], data)
			buffer = buffer[:len(data)]
			s.writeChannel <- buffer

			s.WriteEnqueuedMetric.AddCount(1, now)
			s.WriteEnqueuedBytesMetric.AddCount(uint64(len(buffer)), now)
		} else {
			// Write Queue Full
			s.WritePool.Release(data)
			log.Err(errWriteMessageDiscarded).Str("msg", hex.EncodeToString(data))
			s.WriteQueueFullMetric.AddCount(1, now)
			s.WriteQueueFullBytesMetric.AddCount(uint64(len(data)), now)
		}
	}
}

func (s *SDLCService) logMessage(data []byte, action string, omitMetric *utils.Metric) {
	if s.CsvProvider != nil {
		now := time.Now()
		writer, err := s.CsvProvider.GetWriter()

		if err != nil {
			omitMetric.AddCount(1, now)
			return
		}

		writer.WriteCol(now.Format(utils.DisplayDateTimeMS))
		writer.WriteCol(action)
		writer.WriteCol(hex.EncodeToString(data))
		writer.WriteLn()
	}
}

func (s *SDLCService) WriteLogHeader(
	_ *utils.CSVRollOverFileWriterProvider,
	writer *utils.CSVWriter,
	_ string,
	_ string,
) {
	writer.WriteColsLn("SDLC Action Data", "101")
	writer.WriteColsLn("Radar Vision", "https://radarvision.ai")
	writer.WriteColsLn()
	writer.WriteColsLn("TIMESTAMP", "ACTION", "DATA")
}

func (s *SDLCService) logError(errObj error) {
	s.Error = errObj
	if s.CsvProvider != nil {
		now := time.Now()
		writer, err := s.CsvProvider.GetWriter()

		if err != nil {
			s.OmitLogReadsMetric.AddCount(1, now)
			return
		}

		writer.WriteCol(now.Format(utils.DisplayDateTimeMS))
		writer.WriteCol(errorAction)
		writer.WriteCol(errObj.Error())
		writer.WriteLn()
	}
}
