package portbroker

import (
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type RadarChannel struct {
	IPAddress      utils.IP4
	buffer         [16000]byte
	SegmentCounter uint16
	SegmentTotal   uint16
	SegmentId      uint16
	Stats          RadarStatistics
	Now            time.Time
	msgChannel     chan *RadarMessage
	doneChannel    chan bool
	fixed          utils.FixedBuffer
	DataSlice      []byte
	terminated     bool
	OnTerminate    func(*RadarChannel)

	diagnosticsHandler HandlerForDiagnostics
	objectListHandler  HandlerForObjectList
	statisticsHandler  HandlerForStatistics
	instructionHandler HandlerForInstruction
	pvrHandler         HandlerForPVR
	triggerHandler     HandlerForTrigger
}

func (rc *RadarChannel) Start(radarIP utils.IP4) {
	rc.IPAddress = radarIP
	rc.msgChannel = make(chan *RadarMessage, 5)
	rc.doneChannel = make(chan bool)
	rc.fixed = utils.NewFixedBuffer(rc.buffer[:], 0, 0)

	rc.Stats.Init(radarIP)
	rc.diagnosticsHandler.Init(rc)
	rc.objectListHandler.Init(rc)
	rc.statisticsHandler.Init(rc)
	rc.instructionHandler.Init(rc)
	rc.pvrHandler.Init(rc)
	rc.triggerHandler.Init(rc)

	go rc.execute()
}

func (rc *RadarChannel) Stop() {
	rc.doneChannel <- true

	for !rc.terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

func (rc *RadarChannel) execute() {
	for {
		select {
		case msg := <-rc.msgChannel:
			rc.startMsg(msg)

		case <-rc.doneChannel:
			close(rc.msgChannel)
			close(rc.doneChannel)

			if rc.OnTerminate != nil {
				rc.OnTerminate(rc)
				rc.terminated = true
			}

			return
		}
	}
}

func (rc *RadarChannel) startMsg(msg *RadarMessage) {
	rc.Now = time.Now()
	rc.Stats.Register(RsTotalMessages, rc.Now)

	process := rc.isTransportHeader(msg)
	if process {
		process = rc.consumeData(msg)
	}

	// Process may be false on segmentation, or error
	if process {
		rc.process()
		rc.resetSegmentation()
	}

	// Put the msgChannel back into the pool
	messagePool.Put(msg)
}

// consumeData copies/concats the msg.Buffer to our own Buffer
// by checking whether it is segmented.  consumeData returns true
// the DataSlice is ready for processing, otherwise false
func (rc *RadarChannel) consumeData(msg *RadarMessage) bool {
	th := port.TransportHeaderReader{
		Buffer: msg.Buffer[:msg.BufferLen],
	}

	// The header says the data is segmented
	if th.GetFlags().IsSegmentation() {
		// The data is segmented, so we have to check whether we are currently
		// segmented

		// Check if currently in segmentation mode
		if rc.SegmentId != 0 {
			// Check if segment out of sequence
			if rc.SegmentId != th.GetDataIdentifier() {
				// The segment is out of sequence, so we have to reset
				rc.Stats.Register(RsDiscardedSegment, rc.Now)
				rc.SegmentCounter = 0
				rc.fixed.Reset()
				return false
			} else {
				// The segment is in sequence to append and process if last
				rc.SegmentCounter += 1

				if !rc.fixed.CanWrite(msg.BufferLen) {
					// Segmentation won't fit
					rc.Stats.Register(RsSegmentationBufferOverflow, rc.Now)
					rc.resetSegmentation()
					return false
				} else {
					rc.fixed.WriteBytes(msg.Buffer[th.GetHeaderLength():msg.BufferLen])
					rc.DataSlice = rc.fixed.AsWriteSlice()
					return rc.SegmentCounter == rc.SegmentTotal
				}
			}
		} else {
			// Start segmentation
			rc.SegmentCounter = 1
			rc.SegmentId = th.GetDataIdentifier()
			rc.SegmentTotal = th.GetSegmentation()
			rc.fixed.WriteBytes(msg.Buffer[:msg.BufferLen])
			return false
		}
	} else {
		if rc.SegmentId != 0 {
			// Currently in segmentation mode and received a msgChannel without
			// segmentation, so we register the discarded segment
			rc.Stats.Register(RsDiscardedSegment, rc.Now)
		}
		// The data is not segmented, then reset whatever segment we may have

		// Simply use the data from the messagePool, as it is complete and does not
		// require any buffer copying
		rc.DataSlice = msg.Buffer[:msg.BufferLen]
		return true
	}
}

func (rc *RadarChannel) isTransportHeader(msg *RadarMessage) bool {
	// Validate the header
	th := port.TransportHeaderReader{
		Buffer: msg.Buffer[:msg.BufferLen],
	}

	if err := th.CheckFormat(); err != nil {
		return rc.invalidTransportHeader(err)
	}

	if err := th.CheckCRC(); err != nil {
		return rc.invalidHeaderCRC(err)
	}

	if th.GetProtocolType() != port.PtSmartMicroPort {
		return rc.unsupportedProtocol()
	}

	return true
}

// process assumes that the DataSlice points to a complete port msgChannel
// The source can either be RadarChannel.buffer (due to segmentation) or the
// RadarMessage.Buffer (no segmentation)
func (rc *RadarChannel) process() {
	th := port.TransportHeaderReader{
		Buffer: rc.DataSlice,
	}
	ph := port.PortHeaderReader{
		Buffer:      rc.DataSlice,
		StartOffset: int(th.GetHeaderLength()),
	}

	if err := ph.Check(); err != nil {
		if rc.Stats.Register(RsPortHeaderFormatErr, rc.Now) {
			rc.logError(err)
			return
		}
	}

	pid := ph.GetIdentifier()
	switch pid {
	case port.PiDiagnostics:
		rc.Stats.Register(RsDiagnosticCount, rc.Now)
		rc.diagnosticsHandler.Process(rc.Now, rc.DataSlice)

	case port.PiObjectList:
		rc.Stats.Register(RsObjectListCount, rc.Now)
		rc.objectListHandler.Process(rc.Now, rc.DataSlice)

	case port.PiStatistics:
		rc.Stats.Register(RsStatisticsCount, rc.Now)
		rc.statisticsHandler.Process(rc.Now, rc.DataSlice)

	case port.PiEventTrigger:
		rc.Stats.Register(RsTriggerCount, rc.Now)
		rc.triggerHandler.Process(rc.Now, rc.DataSlice)

	case port.PiPVR:
		rc.Stats.Register(RsPVRCount, rc.Now)
		rc.pvrHandler.Process(rc.Now, rc.DataSlice)

	case port.PiInstruction:
		rc.Stats.Register(RsInstructionCount, rc.Now)
		rc.instructionHandler.Process(rc.Now, rc.DataSlice)

	default:
		rc.Stats.Register(RsUnknownPortIdentifier, rc.Now)
	}
}

func (rc *RadarChannel) invalidTransportHeader(err error) bool {
	if rc.Stats.Register(RsTransportHeaderFormatErr, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) invalidHeaderCRC(err error) bool {
	if rc.Stats.Register(RsTransportHeaderCRCErr, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) unsupportedProtocol() bool {
	if rc.Stats.Register(RsProtocolTypeErr, rc.Now) {
		rc.logError(port.ErrUnsupportedProtocol)
	}
	return false
}

func (rc *RadarChannel) logError(err error) {
	log.Err(err).Msgf("radar: %s", rc.IPAddress)
}

func (rc *RadarChannel) resetSegmentation() {
	rc.SegmentTotal = 0
	rc.SegmentCounter = 0
	rc.SegmentId = 0
	rc.fixed.Reset()
}

func (rc *RadarChannel) SendMessage(msg *RadarMessage) {
	if len(rc.msgChannel) < cap(rc.msgChannel) {
		rc.msgChannel <- msg
	} else {
		// The message gets dropped because the channel queue is full
		rc.Stats.Register(RsMessageDrop, msg.CreateOn)
		messagePool.Put(msg)
	}
}
