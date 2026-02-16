package uartsdlc

import (
	"encoding/hex"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
	"rvpro3/radarvision.com/internal/branding"
	"rvpro3/radarvision.com/utils"
)

var errWriteMessageDiscarded = errors.New("write message discarded")

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
	Serial             SerialConnection
	doneChan           chan bool
	writeChannel       chan []byte
	readBuffer         [1024]byte
	backingBuffer      [2048]byte
	serialBuffer       utils.SerialBuffer
	terminate          bool
	terminateRefCount  atomic.Int32
	Metrics            SDLCServiceMetrics
	WritePool          *SDLCWritePool             `json:"-"`
	OnError            func(*SDLCService, error)  `json:"-"`
	OnTerminate        func(*SDLCService)         `json:"-"`
	OnReadMessage      func(*SDLCService, []byte) `json:"-"`
	OnWriteMessage     func(*SDLCService, []byte) `json:"-"`
	RetrySleepDuration time.Duration
	Error              error
	CsvProvider        *utils.CSVRollOverFileWriterProvider `json:"-"`
}

type SDLCServiceMetrics struct {
	IsWriteEnabled      *utils.Metric
	WriteEnqueued       *utils.Metric
	WriteEnqueuedBytes  *utils.Metric
	WriteDequeued       *utils.Metric
	WriteDequeuedBytes  *utils.Metric
	WriteQueueFull      *utils.Metric
	WriteQueueFullBytes *utils.Metric
	WriteSuccess        *utils.Metric
	WriteSuccessBytes   *utils.Metric
	WriteError          *utils.Metric
	WriteErrorBytes     *utils.Metric
	WriteOmit           *utils.Metric
	WriteOmitBytes      *utils.Metric
	OmitLogWrites       *utils.Metric
	OmitLogReads        *utils.Metric
	ReadBytes           *utils.Metric
	Reads               *utils.Metric
	Pops                *utils.Metric
	PopBytes            *utils.Metric
	utils.MetricsInitMixin
}

func (s *SDLCService) SetupDefaults(config *utils.Settings) {
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

func (s *SDLCService) SetupAndStart(state *utils.State, config *utils.Settings) {
	if !config.GetSettingAsBool(sdlcUARTEnabled) {
		log.Info().Msg("SDLC UART is disabled")
		return
	}

	s.InitFromConfig(config)
	s.Start()

	state.Set(SDLCServiceName, s)
}

func (s *SDLCService) InitFromConfig(config *utils.Settings) {
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
	s.Metrics.InitMetrics(SDLCServiceName, &s.Metrics)
	s.RetrySleepDuration = time.Duration(1) * time.Second
	s.WritePool = NewSDLCWritePool()
	s.Serial.RetryGuard.RetryEvery = 3
	s.serialBuffer.Buffer = s.backingBuffer[:]
	s.serialBuffer.StartDelim = 0x02
	s.serialBuffer.EndDelim = 0x03
	s.doneChan = make(chan bool)
	s.writeChannel = make(chan []byte, 5)
	s.terminateRefCount.Store(2)
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
	s.Metrics.IsWriteEnabled.SetBool(true)

	for !s.terminate {

		if s.Serial.Connect() {
			readSize := s.Serial.Read(s.readBuffer[:])
			if readSize > 0 {
				now := time.Now()
				s.Metrics.Reads.IncAt(1, now)
				s.Metrics.ReadBytes.IncAt(int64(readSize), now)

				if err := s.serialBuffer.Push(s.readBuffer[:readSize]); err != nil {
					s.logError(err)
					if s.OnError != nil {
						s.OnError(s, err)
					}
				} else {
					if readBytes := s.serialBuffer.Pop(); readBytes != nil {
						s.Metrics.Pops.IncAt(1, now)
						s.Metrics.PopBytes.IncAt(int64(len(readBytes)), now)

						s.logMessage(readBytes, readAction, s.Metrics.OmitLogReads)
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
			s.Metrics.IsWriteEnabled.SetBool(false)
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

	if s.Metrics.IsWriteEnabled.GetBool() {
		if s.Serial.Write(data) {
			s.Metrics.WriteSuccess.IncAt(1, now)
			s.Metrics.WriteSuccessBytes.IncAt(int64(len(data)), now)
			s.logMessage(data, writeAction, s.Metrics.OmitLogWrites)
		} else {
			s.Metrics.WriteError.IncAt(1, now)
			s.Metrics.WriteErrorBytes.IncAt(int64(len(data)), now)
		}
	} else {
		s.Metrics.WriteOmit.IncAt(1, now)
		s.Metrics.WriteOmitBytes.IncAt(int64(len(data)), now)
	}

	s.WritePool.Release(data)
	s.Metrics.WriteDequeuedBytes.IncAt(int64(len(data)), now)
	s.Metrics.WriteDequeued.IncAt(1, now)
}

func (s *SDLCService) Write(data []byte) {
	if !s.terminate {
		now := time.Now()

		if len(s.writeChannel) < cap(s.writeChannel) {
			buffer := s.WritePool.Alloc()
			copy(buffer[0:len(data)], data)
			buffer = buffer[:len(data)]
			s.writeChannel <- buffer

			s.Metrics.WriteEnqueued.IncAt(1, now)
			s.Metrics.WriteEnqueuedBytes.IncAt(int64(len(buffer)), now)
		} else {
			// WritePacket Queue Full
			s.WritePool.Release(data)
			log.Err(errWriteMessageDiscarded).Str("msg", hex.EncodeToString(data))
			s.Metrics.WriteQueueFull.IncAt(1, now)
			s.Metrics.WriteQueueFullBytes.IncAt(int64(len(data)), now)
		}
	}
}

func (s *SDLCService) logMessage(data []byte, action string, omitMetric *utils.Metric) {
	if s.CsvProvider != nil {
		now := time.Now()
		writer, err := s.CsvProvider.GetWriter()

		if err != nil {
			omitMetric.IncAt(1, now)
			return
		}

		writer.WriteCol(now.Format(utils.DisplayDateTimeMS))
		writer.WriteCol(action)
		writer.WriteColNL(hex.EncodeToString(data))
	}
}

func (s *SDLCService) WriteLogHeader(
	_ *utils.CSVRollOverFileWriterProvider,
	writer *utils.CSVWriter,
	_ string,
	_ string,
) {
	branding.CSVBranding.WriteTitle(writer, "SDLC Action Value", "101")
	branding.CSVBranding.WriteFeaturesNL(writer)
	writer.WriteColsNL("TIMESTAMP", "ACTION", "DATA")
}

func (s *SDLCService) logError(errObj error) {
	s.Error = errObj
	if s.CsvProvider != nil {
		now := time.Now()
		writer, err := s.CsvProvider.GetWriter()

		if err != nil {
			s.Metrics.OmitLogReads.IncAt(1, now)
			return
		}

		writer.WriteCol(now.Format(utils.DisplayDateTimeMS))
		writer.WriteCol(errorAction)
		writer.WriteColNL(errObj.Error())
	}
}
