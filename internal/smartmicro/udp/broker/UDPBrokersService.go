package broker

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

const UDPBrokersServiceName = "UDP.Brokers.Service"

type UDPBrokersService struct {
	Brokers           []UDPBroker   `json:"Broker"`
	TerminateRefCount atomic.Uint32 `json:"-"`
	workflowBuilder   interfaces.IUDPWorkflowBuilder
}

func (rc *UDPBrokersService) SetupDefaults(config *utils.Settings) {
}

func (rc *UDPBrokersService) getChannelConfig() *servicemodel.Config {
	res, ok := utils.GlobalState.Get(servicemodel.StateName).(*servicemodel.Config)
	if !ok {
		panic("func (rc *UDPBrokersService) getChannelConfig() *servicemodel.Config")
	}
	return res
}

func (rc *UDPBrokersService) SetupAndStart(state *utils.State, _ *utils.Settings) {
	state.Set("UDP.Brokers", rc)
	dataService, ok := state.Get(service.UDPDataServiceName).(*service.UDPData)

	if !ok {
		log.Warn().Msg("UDP Brokers not configured due to no UDP data service...")
		return
	}

	serviceCfg := rc.getChannelConfig()
	rc.InitNoRadars(len(serviceCfg.Radars))
	rc.AttachTo(dataService)

	for index, radarCfg := range serviceCfg.Radars {
		channel := &rc.Brokers[index]
		channel.InitMetrics(radarCfg.GetRadarIP())
		channel.SetupWorkflow(channel, serviceCfg, radarCfg)
	}

	rc.SetupStates(state)

	rc.Start()
}

func (rc *UDPBrokersService) SetupStates(state *utils.State) {
	// Setup Global Trigger Pipeline State
	state.Set(
		triggerpipeline.TriggerPipelineStateName,
		new(triggerpipeline.TriggerPipeline),
	)

	// Setup Global Phase Sate
	state.Set(interfaces.PhaseStateName, new(interfaces.PhaseState))
}

func (rc *UDPBrokersService) GetServiceName() string { return UDPBrokersServiceName }

func (rc *UDPBrokersService) GetServiceNames() []string {
	return nil
}

func (rc *UDPBrokersService) InitNoRadars(numberOfRadars int) {
	rc.Brokers = make([]UDPBroker, numberOfRadars)
}

func (rc *UDPBrokersService) Start() {
	rc.TerminateRefCount.Store(0)

	for index := range rc.Brokers {
		radar := &rc.Brokers[index]
		radar.OnTerminate = rc.OnChannelTerminate
		radar.Run(radar.GetRadarIP())
	}
}

func (rc *UDPBrokersService) Stop() {
	for index := range rc.Brokers {
		radar := &rc.Brokers[index]
		radar.Stop()
	}
}

func (rc *UDPBrokersService) AwaitStop(sleepTime time.Duration) {
	for rc.TerminateRefCount.Load() < 4 {
		time.Sleep(sleepTime * time.Millisecond)
	}
}

func (rc *UDPBrokersService) AttachTo(udp *service.UDPData) {
	udp.OnData = rc.OnData
}

func (rc *UDPBrokersService) OnData(
	dataService *service.UDPData,
	addr net.UDPAddr,
	bytes []byte,
) {
	ip4 := utils.IP4Builder.FromIP(addr.IP, addr.Port)
	radarIndex := rc.getChannelConfig().GetRadarIndex(ip4)

	if radarIndex == -1 {
		dataService.Metrics.UnmappedRadarPacket.Inc(1)
		return
	}

	msg := messagePool.Get().(*UDPMessage)
	msg.BufferLen = len(bytes)
	msg.IPAddress = utils.IP4Builder.FromIP(addr.IP, addr.Port)
	msg.CreateOn = time.Now()
	copy(msg.Buffer[:], bytes)

	radar := &rc.Brokers[radarIndex]
	radar.SendMessage(msg)
}

func (rc *UDPBrokersService) OnChannelTerminate(channel *UDPBroker) {
	rc.TerminateRefCount.Add(1)
}
