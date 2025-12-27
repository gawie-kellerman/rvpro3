package udp

import (
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/instrumentation"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type RadarChannel struct {
	Metrics        *instrumentation.RadarMetrics
	State          RadarState
	IPAddress      utils.IP4
	buffer         [16000]byte
	SegmentCounter uint16
	SegmentTotal   uint16
	SegmentId      uint16
	Now            time.Time
	msgChannel     chan *RadarMessage
	doneChannel    chan bool
	fixed          utils.FixedBuffer
	DataSlice      []byte
	terminated     bool
	OnTerminate    func(*RadarChannel)

	isDone              bool
	DiagnosticsWorkflow interfaces.IUDPWorkflow
	ObjectListWorkflow  interfaces.IUDPWorkflow
	StatisticsWorkflow  interfaces.IUDPWorkflow
	InstructionWorkflow interfaces.IUDPWorkflow
	PvrWorkflow         interfaces.IUDPWorkflow
	TriggersWorkflow    interfaces.IUDPWorkflow
}

func (rc *RadarChannel) Run(radarIP utils.IP4, workflowBuilder interfaces.IUDPWorkflowBuilder) {
	rc.IPAddress = radarIP
	rc.isDone = false
	rc.msgChannel = make(chan *RadarMessage, 5)
	rc.doneChannel = make(chan bool)
	rc.fixed = utils.NewFixedBuffer(rc.buffer[:], 0, 0)

	rc.Metrics.RadarIP = radarIP
	rc.DiagnosticsWorkflow = workflowBuilder.GetDiagnosticsWorkflow(rc)
	rc.ObjectListWorkflow = workflowBuilder.GetObjectListWorkflow(rc)
	rc.StatisticsWorkflow = workflowBuilder.GetStatisticsWorkflow(rc)
	rc.InstructionWorkflow = workflowBuilder.GetInstructionWorkflow(rc)
	rc.PvrWorkflow = workflowBuilder.GetPVRWorkflow(rc)
	rc.TriggersWorkflow = workflowBuilder.GetTriggerWorkflow(rc)

	rc.DiagnosticsWorkflow.SetParent(rc)
	rc.ObjectListWorkflow.SetParent(rc)
	rc.StatisticsWorkflow.SetParent(rc)
	rc.InstructionWorkflow.SetParent(rc)
	rc.PvrWorkflow.SetParent(rc)
	rc.TriggersWorkflow.SetParent(rc)

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
			rc.isDone = true
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
	rc.IncCount(instrumentation.RmtTotalMessagesProcessed, 1, rc.Now)
	rc.IncCount(instrumentation.RmtTotalBytesProcessed, uint64(msg.BufferLen), rc.Now)

	th := port.TransportHeaderReader{}
	process := rc.isTransportHeader(&th, msg)
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

func (rc *RadarChannel) IncCount(metric instrumentation.RadarMetricType, count uint64, time time.Time) bool {
	return rc.Metrics.AddCount(int(metric), count, time)
}

func (rc *RadarChannel) handleSegmentation(msg *RadarMessage, th *port.TransportHeaderReader) bool {
	// The data is segmented, so we have to check whether we are currently
	// segmented

	// Check if currently in segmentation mode
	if rc.SegmentId != 0 {
		// Check if segment out of sequence
		if rc.SegmentId != th.GetDataIdentifier() {
			// The segment is out of sequence, so we have to reset
			rc.IncCount(instrumentation.RmtDiscardedSegment, 1, rc.Now)
			rc.SegmentCounter = 0
			rc.fixed.Reset()
			return false
		}
		// The segment is in sequence to append and process if last
		rc.SegmentCounter += 1

		if !rc.fixed.CanWrite(msg.BufferLen) {
			// Segmentation won't fit
			rc.IncCount(instrumentation.RmtSegmentationBufferOverflow, 1, rc.Now)
			rc.resetSegmentation()
			return false
		}
		rc.fixed.WriteBytes(msg.Buffer[th.GetHeaderLength():msg.BufferLen])
		rc.DataSlice = rc.fixed.AsWriteSlice()
		return rc.SegmentCounter == rc.SegmentTotal
	}
	// Run segmentation
	rc.SegmentCounter = 1
	rc.SegmentId = th.GetDataIdentifier()
	rc.SegmentTotal = th.GetSegmentation()
	rc.fixed.WriteBytes(msg.Buffer[:msg.BufferLen])
	return false
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
		return rc.handleSegmentation(msg, &th)
	}
	if rc.SegmentId != 0 {
		// Currently in segmentation mode and received a msgChannel without
		// segmentation, so we register the discarded segment
		rc.IncCount(instrumentation.RmtDiscardedSegment, 1, rc.Now)
	}
	// The data is not segmented, then reset whatever segment we may have

	// Simply use the data from the messagePool, as it is complete and does not
	// require any buffer copying
	rc.DataSlice = msg.Buffer[:msg.BufferLen]
	return true
}

func (rc *RadarChannel) isTransportHeader(th *port.TransportHeaderReader, msg *RadarMessage) bool {
	// Validate the header
	th.Buffer = msg.Buffer[:msg.BufferLen]

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
		if rc.IncCount(instrumentation.RmtHeaderFormatErr, 1, rc.Now) {
			rc.logError(err)
			return
		}
	}

	pid := ph.GetIdentifier()

	totalMetricID := 0
	minMetricID := 0
	maxMetricID := 0

	switch pid {
	case port.PiDiagnostics:
		rc.Metrics.AddCount(int(instrumentation.RmtDiagnosticProcessed), 1, rc.Now)
		rc.DiagnosticsWorkflow.Process(rc.Now, rc.DataSlice)
		totalMetricID = int(instrumentation.RmtDiagnosticTotalTime)
		minMetricID = int(instrumentation.RmtDiagnosticMinTime)
		maxMetricID = int(instrumentation.RmtDiagnosticMaxTime)

	case port.PiObjectList:
		rc.Metrics.AddCount(int(instrumentation.RmtObjectListProcessed), 1, rc.Now)
		rc.ObjectListWorkflow.Process(rc.Now, rc.DataSlice)
		totalMetricID = int(instrumentation.RmtObjectListTotalTime)
		minMetricID = int(instrumentation.RmtObjectListMinTime)
		maxMetricID = int(instrumentation.RmtObjectListMaxTime)

	case port.PiStatistics:
		rc.Metrics.AddCount(int(instrumentation.RmtStatisticsProcessed), 1, rc.Now)
		rc.StatisticsWorkflow.Process(rc.Now, rc.DataSlice)
		totalMetricID = int(instrumentation.RmtStatisticsTotalTime)
		minMetricID = int(instrumentation.RmtStatisticsMinTime)
		maxMetricID = int(instrumentation.RmtStatisticsMaxTime)

	case port.PiEventTrigger:
		rc.Metrics.AddCount(int(instrumentation.RmtTriggerProcessed), 1, rc.Now)
		rc.TriggersWorkflow.Process(rc.Now, rc.DataSlice)
		totalMetricID = int(instrumentation.RmtTriggerTotalTime)
		minMetricID = int(instrumentation.RmtTriggerMinTime)
		maxMetricID = int(instrumentation.RmtTriggerMaxTime)

	case port.PiPVR:
		rc.Metrics.AddCount(int(instrumentation.RmtPVRProcessed), 1, rc.Now)
		rc.PvrWorkflow.Process(rc.Now, rc.DataSlice)
		totalMetricID = int(instrumentation.RmtPVRTotalTime)
		minMetricID = int(instrumentation.RmtPVRMinTime)
		maxMetricID = int(instrumentation.RmtPVRMaxTime)

	case port.PiInstruction:
		rc.Metrics.AddCount(int(instrumentation.RmtInstructionProcessed), 1, rc.Now)
		rc.InstructionWorkflow.Process(rc.Now, rc.DataSlice)
		totalMetricID = int(instrumentation.RmtInstructionTotalTime)
		minMetricID = int(instrumentation.RmtInstructionMinTime)
		maxMetricID = int(instrumentation.RmtInstructionMaxTime)

	default:
		rc.Metrics.AddCount(int(instrumentation.RmtUnknownPortIdentifier), 1, rc.Now)
	}

	// Time metric not performed for unknown port identifier
	if totalMetricID != 0 {
		duration := time.Since(rc.Now).Milliseconds()
		rc.Metrics.GetRel(totalMetricID).AddCount(uint64(duration), rc.Now)
		rc.Metrics.GetRel(minMetricID).ReplaceMinDuration(duration, rc.Now)
		rc.Metrics.GetRel(maxMetricID).ReplaceMaxDuration(duration, rc.Now)
	}
}

func (rc *RadarChannel) invalidTransportHeader(err error) bool {
	if rc.Metrics.AddCount(int(instrumentation.RmtTransportHeaderFormatErr), 1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) invalidHeaderCRC(err error) bool {
	if rc.Metrics.AddCount(int(instrumentation.RmtTransportHeaderCRCErr), 1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) unsupportedProtocol() bool {
	if rc.Metrics.AddCount(int(instrumentation.RmtProtocolTypeErr), 1, rc.Now) {
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
	if rc.isDone {
		return
	}

	if len(rc.msgChannel) < cap(rc.msgChannel) {
		rc.msgChannel <- msg
	} else {
		// The message gets dropped because the channel queue is full
		rc.logDroppedMessage(msg)
		messagePool.Put(msg)
	}
}

func (rc *RadarChannel) logDroppedMessage(msg *RadarMessage) {
	now := time.Now()

	rc.Metrics.AddCount(int(instrumentation.RmtTotalMessagesDropped), 1, now)

	th := port.TransportHeaderReader{}
	if !rc.isTransportHeader(&th, msg) {
		return
	}

	if th.GetFlags().IsSegmentation() {
		rc.Metrics.AddCount(int(instrumentation.RmtSegmentationDropped), 1, now)
	}

	ph := port.PortHeaderReader{
		Buffer:      rc.DataSlice,
		StartOffset: int(th.GetHeaderLength()),
	}

	pid := ph.GetIdentifier()
	switch pid {
	case port.PiDiagnostics:
		rc.Metrics.AddCount(int(instrumentation.RmtDiagnosticDropped), 1, now)

	case port.PiObjectList:
		rc.Metrics.AddCount(int(instrumentation.RmtObjectListDropped), 1, now)

	case port.PiStatistics:
		rc.Metrics.AddCount(int(instrumentation.RmtStatisticsDropped), 1, now)

	case port.PiEventTrigger:
		rc.Metrics.AddCount(int(instrumentation.RmtTriggerDropped), 1, now)

	case port.PiPVR:
		rc.Metrics.AddCount(int(instrumentation.RmtPVRDropped), 1, now)

	case port.PiInstruction:
		rc.Metrics.AddCount(int(instrumentation.RmtInstructionDropped), 1, now)

	default:
		rc.Metrics.AddCount(int(instrumentation.RmtUnknownDropped), 1, rc.Now)
	}

}
