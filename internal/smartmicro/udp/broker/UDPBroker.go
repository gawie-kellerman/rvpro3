package broker

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/models/servicemodel"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/objectlist"
	pvr "rvpro3/radarvision.com/internal/smartmicro/udp/activity/pvr"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/statistics"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/trigger"
	"rvpro3/radarvision.com/utils"
)

type UDPBroker struct {
	buffer           [16000]byte
	fixed            utils.FixedBuffer
	terminated       bool
	isDone           bool
	msgChannel       chan *UDPMessage
	doneChannel      chan bool
	State            RadarState
	IPAddress        utils.IP4
	SegmentCounter   uint16
	SegmentTotal     uint16
	SegmentId        uint16
	Now              time.Time
	DataSlice        []byte           `json:"-"`
	OnTerminate      func(*UDPBroker) `json:"-"`
	Metrics          UDPBrokerMetrics `json:"-"`
	FailSafePipeline triggerpipeline.RadarFailsafeItem
	Executor         WorkflowExecutor
}

type UDPBrokerMetrics struct {
	ReceivedCount            *utils.Metric
	ReceivedBytes            *utils.Metric
	SendMessageDrops         *utils.Metric
	TransportHeaderFormatErr *utils.Metric
	TransportHeaderCrcErr    *utils.Metric
	ProtocolTypeErr          *utils.Metric
	DiscardSegmentErr        *utils.Metric
	UnknownPortErr           *utils.Metric
	SegmentBufferOverflowErr *utils.Metric
	PortHeaderFormatErr      *utils.Metric
	utils.MetricsInitMixin
}

func (rc *UDPBroker) SetupDefaults(config *utils.Settings) {
	config.GetOrPutBool("udp.verbose.trigger", false)
	config.GetOrPutBool("udp.verbose.statistics", false)
	config.GetOrPutBool("udp.verbose.objectlist", false)
	config.GetOrPutBool("udp.verbose.pvr", false)
}

func (rc *UDPBroker) SetupAndStart(state *utils.State, config *utils.Settings) {
	//radars := config.GetSettingAsSplit(radarChannelSupportedRadars, ",")
	//noRadars := len(radars)
}

func (rc *UDPBroker) GetServiceName() string {
	return "Broker." + rc.IPAddress.String() + ".Service"
}

func (rc *UDPBroker) GetServiceNames() []string {
	return nil
}

func (rc *UDPBroker) InitMetrics(ip utils.IP4) {
	rc.IPAddress = ip
	sectionName := fmt.Sprintf("UDP.Broker.[%s]", ip)
	rc.Metrics.InitMetrics(sectionName, &rc.Metrics)
	rc.Executor.Init(ip)
}

func (rc *UDPBroker) GetRadarIP() utils.IP4 {
	return rc.IPAddress
}

func (rc *UDPBroker) Run(radarIP utils.IP4) {
	rc.InitMetrics(radarIP)
	rc.IPAddress = radarIP
	rc.isDone = false
	rc.msgChannel = make(chan *UDPMessage, 5)
	rc.doneChannel = make(chan bool)
	rc.fixed = utils.NewFixedBuffer(rc.buffer[:], 0, 0)

	rc.SetupFailSafe()

	go rc.execute()
}

func (rc *UDPBroker) Stop() {
	rc.doneChannel <- true

	for !rc.terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

func (rc *UDPBroker) execute() {
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

func (rc *UDPBroker) startMsg(msg *UDPMessage) {
	rc.Now = time.Now()
	rc.Metrics.ReceivedCount.IncAt(1, rc.Now)
	rc.Metrics.ReceivedBytes.IncAt(int64(msg.BufferLen), rc.Now)

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

	// Put the UDPMessage back into the pool
	messagePool.Put(msg)
}

func (rc *UDPBroker) handleSegmentation(msg *UDPMessage, th *port.TransportHeaderReader) bool {
	// The data is segmented, so we have to check whether we are currently
	// segmented

	// Check if currently in segmentation mode
	if rc.SegmentId != 0 {
		// Check if segment out of sequence
		if rc.SegmentId != th.GetDataIdentifier() {
			// The segment is out of sequence, so we have to reset
			rc.Metrics.DiscardSegmentErr.IncAt(1, rc.Now)
			rc.SegmentCounter = 0
			rc.fixed.Reset()
			return false
		}
		// The segment is in sequence to append and process if last
		rc.SegmentCounter += 1

		if !rc.fixed.CanWrite(msg.BufferLen) {
			// Segmentation won't fit
			rc.Metrics.SegmentBufferOverflowErr.IncAt(1, rc.Now)
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
func (rc *UDPBroker) consumeData(msg *UDPMessage) bool {
	th := port.TransportHeaderReader{
		Buffer: msg.Buffer[:msg.BufferLen],
	}

	// The header says the data is segmented
	if th.GetFlags().IsSegmentation() {
		return rc.handleSegmentation(msg, &th)
	}
	if rc.SegmentId != 0 {
		// Currently in segmentation mode and received a message without
		// segmentation, so we register the discarded segment
		rc.Metrics.DiscardSegmentErr.IncAt(1, rc.Now)
	}
	// The data is not segmented, then reset whatever segment we may have

	// Simply use the data from the messagePool, as it is complete and does not
	// require any buffer copying
	rc.DataSlice = msg.Buffer[:msg.BufferLen]
	return true
}

func (rc *UDPBroker) isTransportHeader(th *port.TransportHeaderReader, msg *UDPMessage) bool {
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

// process assumes that the DataSlice points to a complete port message
// The source can either be UDPBroker.buffer (due to segmentation) or the
// UDPMessage.Buffer (no segmentation)
func (rc *UDPBroker) process() {
	th := port.TransportHeaderReader{
		Buffer: rc.DataSlice,
	}
	ph := port.PortHeaderReader{
		Buffer:      rc.DataSlice,
		StartOffset: int(th.GetHeaderLength()),
	}

	if err := ph.Check(); err != nil {
		if rc.Metrics.PortHeaderFormatErr.IncAt(1, rc.Now) {
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

func (rc *UDPBroker) invalidTransportHeader(err error) bool {
	if rc.Metrics.TransportHeaderFormatErr.IncAt(1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *UDPBroker) invalidHeaderCRC(err error) bool {
	if rc.Metrics.TransportHeaderCrcErr.IncAt(1, rc.Now) {
		rc.logError(err)
	}

	return false
}

func (rc *UDPBroker) unsupportedProtocol() bool {
	if rc.Metrics.UnknownPortErr.IncAt(1, rc.Now) {
		rc.logError(port.ErrUnsupportedProtocol)
	}
	return false
}

func (rc *UDPBroker) logError(err error) {
	log.Err(err).Msgf("radar: %s", rc.IPAddress)
}

func (rc *UDPBroker) resetSegmentation() {
	rc.SegmentTotal = 0
	rc.SegmentCounter = 0
	rc.SegmentId = 0
	rc.fixed.Reset()
}

func (rc *UDPBroker) SendMessage(msg *UDPMessage) {
	if rc.isDone {
		return
	}

	if len(rc.msgChannel) < cap(rc.msgChannel) {
		rc.msgChannel <- msg
	} else {
		// The message gets dropped because the channel queue is full
		rc.Metrics.SendMessageDrops.Inc(1)
		messagePool.Put(msg)
	}
}

//func (rc *UDPBroker) logDroppedMessage(msg *UDPMessage) {
//	rc.Metrics.SendMessageDrops.IncAt(1, rc.Now)
//
//	th := port.TransportHeaderReader{}
//	if !rc.isTransportHeader(&th, msg) {
//		return
//	}
//
//	if th.GetFlags().IsSegmentation() {
//		rc.Metrics.DiscardSegmentErr.IncAt(1, rc.Now)
//	}
//
//	ph := port.PortHeaderReader{
//		Buffer:      rc.DataSlice,
//		StartOffset: int(th.GetHeaderLength()),
//	}
//
//	pid := ph.GetIdentifier()
//	rc.Executor.Drop(rc.Now, uint32(pid), rc.DataSlice)
//
//}

func (rc *UDPBroker) SetupFailSafe() {
	//pipeline := triggerpipeline.GetTriggerPipeline()
	//rc.FailSafePipeline.SetChannels.Lo = 15
}

func (rc *UDPBroker) SetupWorkflow(
	channel *UDPBroker,
	serviceCfg *servicemodel.Config,
	radarCfg *servicemodel.Radar,
) {
	var flow interfaces.IUDPWorkflow

	if utils.GlobalSettings.GetOrPutBool("udp.verbose.trigger", false) {
		flow = rc.Executor.Workflow(port.PiEventTrigger)
		flow.AddActivity(&trigger.LogToConsoleActivity{})
	}

	if utils.GlobalSettings.GetOrPutBool("udp.verbose.objectlist", false) {
		flow = rc.Executor.Workflow(port.PiObjectList)
		flow.AddActivity(&objectlist.LogToConsoleActivity{})
	}

	if utils.GlobalSettings.GetOrPutBool("udp.verbose.statistics", false) {
		flow = rc.Executor.Workflow(port.PiStatistics)
		flow.AddActivity(&statistics.LogToConsoleActivity{})
	}

	if utils.GlobalSettings.GetOrPutBool("udp.verbose.pvr", false) {
		flow = rc.Executor.Workflow(port.PiPVR)
		flow.AddActivity(&pvr.LogToConsoleActivity{})
	}
}
