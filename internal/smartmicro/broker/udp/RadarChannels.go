package udp

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/models/servicemodel"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
)

const RadarChannelsServiceName = "Radar.Channels.Service"

type RadarChannels struct {
	Radar             []RadarChannel
	TerminateRefCount atomic.Uint32
	workflowBuilder   interfaces.IUDPWorkflowBuilder
}

func (rc *RadarChannels) SetupDefaults(config *utils.Settings) {
}

func (rc *RadarChannels) getChannelConfig() *servicemodel.Config {
	res, ok := utils.GlobalState.Get(servicemodel.StateName).(*servicemodel.Config)
	if !ok {
		panic("func (rc *RadarChannels) getChannelConfig() *servicemodel.Config")
	}
	return res
}

func (rc *RadarChannels) SetupAndStart(state *utils.State, _ *utils.Settings) {
	dataService, ok := state.Get(service.UDPDataServiceName).(*service.UDPData)

	if !ok {
		log.Warn().Msg("Radar Channels not configured due to no UDP data service...")
		return
	}

	serviceCfg := rc.getChannelConfig()
	rc.InitNoRadars(len(serviceCfg.Radars))
	rc.AttachTo(dataService)

	for index, radarCfg := range serviceCfg.Radars {
		channel := &rc.Radar[index]
		channel.InitMetrics(radarCfg.GetRadarIP())
		channel.SetupWorkflow(channel, serviceCfg, radarCfg)
	}

	rc.SetupStates(state)

	rc.Start()
}

func (rc *RadarChannels) SetupStates(state *utils.State) {
	// Setup Global Trigger Pipeline State
	state.Set(
		triggerpipeline.TriggerPipelineStateName,
		new(triggerpipeline.TriggerPipeline),
	)

	// Setup Global Phase Sate
	state.Set(interfaces.PhaseStateName, new(interfaces.PhaseState))
}

func (rc *RadarChannels) GetServiceName() string { return RadarChannelsServiceName }

func (rc *RadarChannels) GetServiceNames() []string {
	return nil
}

func (rc *RadarChannels) InitNoRadars(numberOfRadars int) {
	rc.Radar = make([]RadarChannel, numberOfRadars)
}

func (rc *RadarChannels) Start() {
	rc.TerminateRefCount.Store(0)

	for index := range rc.Radar {
		radar := &rc.Radar[index]
		radar.OnTerminate = rc.OnChannelTerminate
		radar.Run(radar.GetRadarIP())
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
	radarIndex := rc.getChannelConfig().GetRadarIndex(ip4)

	if radarIndex == -1 {
		dataService.Metrics.UnmappedRadarPacket.Inc(time.Now())
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
