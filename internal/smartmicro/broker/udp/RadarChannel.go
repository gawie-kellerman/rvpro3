package udp

import (
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
)

const radarChannelSupportedRadars = "UDP.Supported.Radars"

type RadarChannel struct {
	MetricsAt                      string
	State                          RadarState
	IPAddress                      utils.IP4
	buffer                         [16000]byte
	SegmentCounter                 uint16
	SegmentTotal                   uint16
	SegmentId                      uint16
	Now                            time.Time
	msgChannel                     chan *RadarMessage
	doneChannel                    chan bool
	fixed                          utils.FixedBuffer
	DataSlice                      []byte
	terminated                     bool
	OnTerminate                    func(*RadarChannel)
	isDone                         bool
	DiagnosticsWorkflow            interfaces.IUDPWorkflow
	ObjectListWorkflow             interfaces.IUDPWorkflow
	StatisticsWorkflow             interfaces.IUDPWorkflow
	InstructionWorkflow            interfaces.IUDPWorkflow
	PvrWorkflow                    interfaces.IUDPWorkflow
	TriggersWorkflow               interfaces.IUDPWorkflow
	totalMessagesProcessedMetric   *utils.Metric
	totalMessagesDroppedMetric     *utils.Metric
	totalBytesProcessedMetric      *utils.Metric
	objListMetric                  messageMetric
	statisticsMetric               messageMetric
	triggersMetric                 messageMetric
	pvrMetric                      messageMetric
	instructionMetric              messageMetric
	diagMetric                     messageMetric
	transportHeaderFormatErrMetric *utils.Metric
	transportHeaderCrcErrMetric    *utils.Metric
	protocolTypeErrMetric          *utils.Metric
	discardSegmentErrMetric        *utils.Metric
	unknownPortErrMetric           *utils.Metric
	segmentBufferOverflowErrMetric *utils.Metric
	portHeaderFormatErrMetric      *utils.Metric
	unknownDroppedMetric           *utils.Metric
	messageProcessedMetric         *utils.Metric
	FailSafePipeline               triggerpipeline.RadarFailsafeItem
}

type messageMetric struct {
	processed *utils.Metric
	dropped   *utils.Metric
	totalTime *utils.Metric
	minTime   *utils.Metric
	maxTime   *utils.Metric
}

func (rc *RadarChannel) SetupDefaults(config *utils.Config) {
}

func (rc *RadarChannel) SetupRunnable(state *utils.State, config *utils.Config) {
	//radars := config.GetSettingAsSplit(radarChannelSupportedRadars, ",")
	//noRadars := len(radars)
}

func (rc *RadarChannel) GetServiceName() string {
	return "Radar." + rc.IPAddress.String() + ".Service"
}

func (rc *RadarChannel) GetServiceNames() []string {
	return nil
}

func (rc *RadarChannel) InitMetrics(ip4 utils.IP4) {
	rc.MetricsAt = interfaces.MetricName.GetUDPRadarMetric(ip4)
	gm := &utils.GlobalMetrics
	rc.totalMessagesProcessedMetric = gm.U64(rc.MetricsAt, "Total Messages processed")
	rc.totalMessagesDroppedMetric = gm.U64(rc.MetricsAt, "Total Messages dropped")
	rc.totalBytesProcessedMetric = gm.U64(rc.MetricsAt, "Total Bytes processed")

	rc.objListMetric.processed = gm.U64(rc.MetricsAt, "Object List processed")
	rc.objListMetric.dropped = gm.U64(rc.MetricsAt, "Object List dropped")
	rc.objListMetric.totalTime = gm.U64(rc.MetricsAt, "Object List totalTime")
	rc.objListMetric.minTime = gm.U64(rc.MetricsAt, "Object List min time")
	rc.objListMetric.maxTime = gm.U64(rc.MetricsAt, "Object List max time")

	rc.statisticsMetric.processed = gm.U64(rc.MetricsAt, "Statistics processed")
	rc.statisticsMetric.dropped = gm.U64(rc.MetricsAt, "Statistics dropped")
	rc.statisticsMetric.totalTime = gm.U64(rc.MetricsAt, "Statistics total time")
	rc.statisticsMetric.minTime = gm.U64(rc.MetricsAt, "Statistics min time")
	rc.statisticsMetric.maxTime = gm.U64(rc.MetricsAt, "Statistics max time")

	rc.triggersMetric.processed = gm.U64(rc.MetricsAt, "Triggers processed")
	rc.triggersMetric.dropped = gm.U64(rc.MetricsAt, "Triggers dropped")
	rc.triggersMetric.totalTime = gm.U64(rc.MetricsAt, "Triggers total time")
	rc.triggersMetric.minTime = gm.U64(rc.MetricsAt, "Triggers min time")
	rc.triggersMetric.maxTime = gm.U64(rc.MetricsAt, "Triggers max time")

	rc.pvrMetric.processed = gm.U64(rc.MetricsAt, "PVR processed")
	rc.pvrMetric.dropped = gm.U64(rc.MetricsAt, "PVR dropped")
	rc.pvrMetric.totalTime = gm.U64(rc.MetricsAt, "PVR total time")
	rc.pvrMetric.minTime = gm.U64(rc.MetricsAt, "PVR min time")
	rc.pvrMetric.maxTime = gm.U64(rc.MetricsAt, "PVR max time")

	rc.instructionMetric.processed = gm.U64(rc.MetricsAt, "Instructions processed")
	rc.instructionMetric.dropped = gm.U64(rc.MetricsAt, "Instructions dropped")
	rc.instructionMetric.totalTime = gm.U64(rc.MetricsAt, "Instructions total time")
	rc.instructionMetric.minTime = gm.U64(rc.MetricsAt, "Instructions min time")
	rc.instructionMetric.maxTime = gm.U64(rc.MetricsAt, "Instructions max time")

	rc.diagMetric.processed = gm.U64(rc.MetricsAt, "Diagnostics processed")
	rc.diagMetric.dropped = gm.U64(rc.MetricsAt, "Diagnostics dropped")
	rc.diagMetric.totalTime = gm.U64(rc.MetricsAt, "Diagnostics total time")
	rc.diagMetric.minTime = gm.U64(rc.MetricsAt, "Diagnostics min time")
	rc.diagMetric.maxTime = gm.U64(rc.MetricsAt, "Diagnostics max time")

	rc.transportHeaderFormatErrMetric = gm.U64(rc.MetricsAt, "Error: Transport header format")
	rc.transportHeaderCrcErrMetric = gm.U64(rc.MetricsAt, "Error: Transport header crc")
	rc.portHeaderFormatErrMetric = gm.U64(rc.MetricsAt, "Error: Port header format")
	rc.protocolTypeErrMetric = gm.U64(rc.MetricsAt, "Error: Protocol type")
	rc.discardSegmentErrMetric = gm.U64(rc.MetricsAt, "Error: Discard segment err")
	rc.segmentBufferOverflowErrMetric = gm.U64(rc.MetricsAt, "Error: Segment buffer overflow")
	rc.unknownPortErrMetric = gm.U64(rc.MetricsAt, "Error: Unknown port")

	rc.unknownDroppedMetric = gm.U64(rc.MetricsAt, "Error: Unknown dropped")
	rc.messageProcessedMetric = gm.U64(rc.MetricsAt, interfaces.MessageProcessedOnMetricName)
}

func (rc *RadarChannel) GetRadarIP() utils.IP4 {
	return rc.IPAddress
}

func (rc *RadarChannel) Run(radarIP utils.IP4, workflowBuilder interfaces.IUDPWorkflowBuilder) {
	rc.InitMetrics(radarIP)
	rc.IPAddress = radarIP
	rc.isDone = false
	rc.msgChannel = make(chan *RadarMessage, 5)
	rc.doneChannel = make(chan bool)
	rc.fixed = utils.NewFixedBuffer(rc.buffer[:], 0, 0)

	rc.DiagnosticsWorkflow = workflowBuilder.GetDiagnosticsWorkflow(rc)
	rc.ObjectListWorkflow = workflowBuilder.GetObjectListWorkflow(rc)
	rc.StatisticsWorkflow = workflowBuilder.GetStatisticsWorkflow(rc)
	rc.InstructionWorkflow = workflowBuilder.GetInstructionWorkflow(rc)
	rc.PvrWorkflow = workflowBuilder.GetPVRWorkflow(rc)
	rc.TriggersWorkflow = workflowBuilder.GetTriggerWorkflow(rc)

	rc.DiagnosticsWorkflow.Init(rc)
	rc.ObjectListWorkflow.Init(rc)
	rc.StatisticsWorkflow.Init(rc)
	rc.InstructionWorkflow.Init(rc)
	rc.PvrWorkflow.Init(rc)
	rc.TriggersWorkflow.Init(rc)

	rc.SetupFailSafe()

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
	rc.totalMessagesProcessedMetric.Inc(rc.Now)
	rc.totalBytesProcessedMetric.Add(msg.BufferLen, rc.Now)

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

func (rc *RadarChannel) handleSegmentation(msg *RadarMessage, th *port.TransportHeaderReader) bool {
	// The data is segmented, so we have to check whether we are currently
	// segmented

	// Check if currently in segmentation mode
	if rc.SegmentId != 0 {
		// Check if segment out of sequence
		if rc.SegmentId != th.GetDataIdentifier() {
			// The segment is out of sequence, so we have to reset
			rc.discardSegmentErrMetric.Inc(rc.Now)
			rc.SegmentCounter = 0
			rc.fixed.Reset()
			return false
		}
		// The segment is in sequence to append and process if last
		rc.SegmentCounter += 1

		if !rc.fixed.CanWrite(msg.BufferLen) {
			// Segmentation won't fit
			rc.segmentBufferOverflowErrMetric.Inc(rc.Now)
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
		rc.discardSegmentErrMetric.Inc(rc.Now)
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
		if rc.portHeaderFormatErrMetric.AddCount(1, rc.Now) {
			rc.logError(err)
			return
		}
	}

	rc.messageProcessedMetric.SetTime(rc.Now)

	pid := ph.GetIdentifier()
	var msgMetric *messageMetric

	switch pid {
	case port.PiDiagnostics:
		rc.DiagnosticsWorkflow.Process(rc.Now, rc.DataSlice)
		msgMetric = &rc.diagMetric

	case port.PiObjectList:
		rc.ObjectListWorkflow.Process(rc.Now, rc.DataSlice)
		msgMetric = &rc.objListMetric

	case port.PiStatistics:
		rc.StatisticsWorkflow.Process(rc.Now, rc.DataSlice)
		msgMetric = &rc.statisticsMetric

	case port.PiEventTrigger:
		rc.TriggersWorkflow.Process(rc.Now, rc.DataSlice)
		msgMetric = &rc.triggersMetric

	case port.PiPVR:
		rc.PvrWorkflow.Process(rc.Now, rc.DataSlice)
		msgMetric = &rc.pvrMetric

	case port.PiInstruction:
		rc.InstructionWorkflow.Process(rc.Now, rc.DataSlice)
		msgMetric = &rc.instructionMetric

	default:
		rc.unknownPortErrMetric.Inc(rc.Now)
	}

	// Time metric not performed for unknown port identifier
	if msgMetric != nil {
		duration := time.Since(rc.Now).Milliseconds()
		msgMetric.processed.Inc(rc.Now)
		msgMetric.totalTime.AddCount(uint64(duration), rc.Now)
		msgMetric.minTime.ReplaceMinDuration(duration, rc.Now)
		msgMetric.maxTime.ReplaceMaxDuration(duration, rc.Now)
	}
}

func (rc *RadarChannel) invalidTransportHeader(err error) bool {
	if rc.transportHeaderFormatErrMetric.AddCount(1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) invalidHeaderCRC(err error) bool {
	if rc.transportHeaderCrcErrMetric.AddCount(1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) unsupportedProtocol() bool {
	if rc.unknownPortErrMetric.AddCount(1, rc.Now) {
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
	rc.totalMessagesDroppedMetric.Inc(rc.Now)

	th := port.TransportHeaderReader{}
	if !rc.isTransportHeader(&th, msg) {
		return
	}

	if th.GetFlags().IsSegmentation() {
		rc.discardSegmentErrMetric.Inc(rc.Now)
	}

	ph := port.PortHeaderReader{
		Buffer:      rc.DataSlice,
		StartOffset: int(th.GetHeaderLength()),
	}

	pid := ph.GetIdentifier()
	switch pid {
	case port.PiDiagnostics:
		rc.diagMetric.dropped.Inc(rc.Now)

	case port.PiObjectList:
		rc.objListMetric.dropped.Inc(rc.Now)

	case port.PiStatistics:
		rc.statisticsMetric.dropped.Inc(rc.Now)

	case port.PiEventTrigger:
		rc.triggersMetric.dropped.Inc(rc.Now)

	case port.PiPVR:
		rc.pvrMetric.dropped.Inc(rc.Now)

	case port.PiInstruction:
		rc.instructionMetric.dropped.Inc(rc.Now)

	default:
		rc.unknownPortErrMetric.Inc(rc.Now)
	}
}

func (rc *RadarChannel) SetupFailSafe() {
	pipeline := triggerpipeline.GetTriggerPipeline()
	rc.FailSafePipeline.SetChannels.Lo = 15
}
