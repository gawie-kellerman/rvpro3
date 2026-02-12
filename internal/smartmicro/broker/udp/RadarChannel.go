package udp

import (
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/models/servicemodel"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
)

type RadarChannel struct {
	State            RadarState
	IPAddress        utils.IP4
	buffer           [16000]byte
	SegmentCounter   uint16
	SegmentTotal     uint16
	SegmentId        uint16
	Now              time.Time
	msgChannel       chan *RadarMessage
	doneChannel      chan bool
	fixed            utils.FixedBuffer
	DataSlice        []byte
	terminated       bool
	OnTerminate      func(*RadarChannel)
	isDone           bool
	Metrics          radarChannelMetric
	FailSafePipeline triggerpipeline.RadarFailsafeItem
	Executor         WorkflowExecutor
}

type radarChannelMetric struct {
	MetricsAt        string
	udpsReceived     *utils.Metric
	udpReceivedBytes *utils.Metric

	totalMessagesDroppedMetric     *utils.Metric
	transportHeaderFormatErrMetric *utils.Metric
	transportHeaderCrcErrMetric    *utils.Metric
	protocolTypeErrMetric          *utils.Metric
	discardSegmentErrMetric        *utils.Metric
	unknownPortErrMetric           *utils.Metric
	segmentBufferOverflowErrMetric *utils.Metric
	portHeaderFormatErrMetric      *utils.Metric
	unknownDroppedMetric           *utils.Metric
}

func (rc *RadarChannel) SetupDefaults(config *utils.Settings) {
}

func (rc *RadarChannel) SetupAndStart(state *utils.State, config *utils.Settings) {
	//radars := config.GetSettingAsSplit(radarChannelSupportedRadars, ",")
	//noRadars := len(radars)
}

func (rc *RadarChannel) GetServiceName() string {
	return "Radar." + rc.IPAddress.String() + ".Service"
}

func (rc *RadarChannel) GetServiceNames() []string {
	return nil
}

func (m *radarChannelMetric) Init(ip4 utils.IP4) {
	m.MetricsAt = interfaces.MetricName.GetUDPRadarMetric(ip4)
	gm := &utils.GlobalMetrics
	m.udpsReceived = gm.U64(m.MetricsAt, "Total Messages processed")
	m.totalMessagesDroppedMetric = gm.U64(m.MetricsAt, "Total Messages dropped")
	m.udpReceivedBytes = gm.U64(m.MetricsAt, "Total Bytes processed")

	m.transportHeaderFormatErrMetric = gm.U64(m.MetricsAt, "Error: Transport header format")
	m.transportHeaderCrcErrMetric = gm.U64(m.MetricsAt, "Error: Transport header crc")
	m.portHeaderFormatErrMetric = gm.U64(m.MetricsAt, "Error: Port header format")
	m.protocolTypeErrMetric = gm.U64(m.MetricsAt, "Error: Protocol type")
	m.discardSegmentErrMetric = gm.U64(m.MetricsAt, "Error: Discard segment err")
	m.segmentBufferOverflowErrMetric = gm.U64(m.MetricsAt, "Error: Segment buffer overflow")
	m.unknownPortErrMetric = gm.U64(m.MetricsAt, "Error: Unknown port")

	m.unknownDroppedMetric = gm.U64(m.MetricsAt, "Error: Unknown dropped")
}

func (rc *RadarChannel) InitMetrics(ip utils.IP4) {
	rc.IPAddress = ip
	rc.Metrics.Init(ip)
}

func (rc *RadarChannel) GetRadarIP() utils.IP4 {
	return rc.IPAddress
}

func (rc *RadarChannel) Run(radarIP utils.IP4) {
	rc.InitMetrics(radarIP)
	rc.IPAddress = radarIP
	rc.isDone = false
	rc.msgChannel = make(chan *RadarMessage, 5)
	rc.doneChannel = make(chan bool)
	rc.fixed = utils.NewFixedBuffer(rc.buffer[:], 0, 0)

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
	rc.Metrics.udpsReceived.Inc(rc.Now)
	rc.Metrics.udpReceivedBytes.Add(msg.BufferLen, rc.Now)

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

	// Put the RadarMessage back into the pool
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
			rc.Metrics.discardSegmentErrMetric.Inc(rc.Now)
			rc.SegmentCounter = 0
			rc.fixed.Reset()
			return false
		}
		// The segment is in sequence to append and process if last
		rc.SegmentCounter += 1

		if !rc.fixed.CanWrite(msg.BufferLen) {
			// Segmentation won't fit
			rc.Metrics.segmentBufferOverflowErrMetric.Inc(rc.Now)
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
		rc.Metrics.discardSegmentErrMetric.Inc(rc.Now)
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
		if rc.Metrics.portHeaderFormatErrMetric.AddCount(1, rc.Now) {
			rc.logError(err)
			return
		}
	}

	rc.Executor.Execute(
		rc.Now,
		uint32(ph.GetIdentifier()),
		rc.DataSlice,
	)
}

func (rc *RadarChannel) invalidTransportHeader(err error) bool {
	if rc.Metrics.transportHeaderFormatErrMetric.AddCount(1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) invalidHeaderCRC(err error) bool {
	if rc.Metrics.transportHeaderCrcErrMetric.AddCount(1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *RadarChannel) unsupportedProtocol() bool {
	if rc.Metrics.unknownPortErrMetric.AddCount(1, rc.Now) {
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
	rc.Metrics.totalMessagesDroppedMetric.Inc(rc.Now)

	th := port.TransportHeaderReader{}
	if !rc.isTransportHeader(&th, msg) {
		return
	}

	if th.GetFlags().IsSegmentation() {
		rc.Metrics.discardSegmentErrMetric.Inc(rc.Now)
	}

	ph := port.PortHeaderReader{
		Buffer:      rc.DataSlice,
		StartOffset: int(th.GetHeaderLength()),
	}

	pid := ph.GetIdentifier()
	rc.Executor.Drop(rc.Now, uint32(pid), rc.DataSlice)

}

func (rc *RadarChannel) SetupFailSafe() {
	//pipeline := triggerpipeline.GetTriggerPipeline()
	//rc.FailSafePipeline.SetChannels.Lo = 15
}

func (rc *RadarChannel) SetupWorkflow(
	channel *RadarChannel,
	serviceCfg *servicemodel.Config,
	radarCfg *servicemodel.Radar,
) {

}
