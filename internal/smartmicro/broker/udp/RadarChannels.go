package udp

import (
	"net"
	"sync/atomic"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/instrumentation"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

type RadarChannels struct {
	Radar             []RadarChannel
	TerminateRefCount atomic.Uint32
}

func (rc *RadarChannels) Init(numberOfRadars int) {
	rc.Radar = make([]RadarChannel, numberOfRadars)
}

func (rc *RadarChannels) Start(workflowBuilder interfaces.IUDPWorkflowBuilder) {
	rc.TerminateRefCount.Store(0)

	for index := range rc.Radar {
		radar := &rc.Radar[index]
		radar.Metrics = instrumentation.GlobalRadarMetrics.ByIndex(index)
		radar.OnTerminate = rc.OnChannelTerminate
		radar.Run(utils.RadarIPOf(index), workflowBuilder)
	}
}

func (rc *RadarChannels) Stop() {
	for index := range rc.Radar {
		radar := &rc.Radar[index]
		radar.Stop()
	}
}

func (rc *RadarChannels) AwaitStop(sleepTime time.Duration) {
	for rc.TerminateRefCount.Load() < 4 {
		time.Sleep(sleepTime * time.Millisecond)
	}
}

func (rc *RadarChannels) AttachTo(udp *service.UDPData) {
	udp.OnData = rc.OnData
}

func (rc *RadarChannels) OnData(
	dataService *service.UDPData,
	addr net.UDPAddr,
	bytes []byte,
) {
	ip4 := utils.IP4Builder.FromIP(addr.IP, addr.Port)
	radarIndex := utils.RadarIndexOf(ip4.ToU32())

	if radarIndex == -1 {
		dataService.CountMetric(instrumentation.UDPMetricIncorrectRadar, 1)
		return
	}

	msg := messagePool.Get().(*RadarMessage)
	msg.BufferLen = len(bytes)
	msg.IPAddress = utils.IP4Builder.FromIP(addr.IP, addr.Port)
	msg.CreateOn = time.Now()
	copy(msg.Buffer[:], bytes)

	radar := &rc.Radar[radarIndex]
	radar.SendMessage(msg)
}

func (rc *RadarChannels) OnChannelTerminate(channel *RadarChannel) {
	rc.TerminateRefCount.Add(1)
}
