package udp

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

const RadarChannelsServiceName = "RadarChannelsService"

const defaultTriggerPath = "Channel.TriggerPath"
const defaultStatisticsPath = "Channel.StatisticsPath"
const defaultObjectListPath = "Channel.ObjectListPath"

type RadarChannels struct {
	Radar             []RadarChannel
	TerminateRefCount atomic.Uint32
	workflowBuilder   interfaces.IUDPWorkflowBuilder
}

func NewRadarChannels(workflow interfaces.IUDPWorkflowBuilder) *RadarChannels {
	return &RadarChannels{
		workflowBuilder: workflow,
	}
}

func (rc *RadarChannels) SetupDefaults(config *utils.Config) {
	config.SetSettingAsStr(
		radarChannelSupportedRadars,
		"192.168.11.12:55555,192.168.11.13:55555,192.168.11.14:55555,192.168.11.15:55555",
	)

	utils.GlobalConfig.SetDefault(defaultTriggerPath, "/media/SDLOGS/logs/sensor/{sensor-host}/trigger/trigger-{datetime}.csv")
	utils.GlobalConfig.SetDefault(defaultStatisticsPath, "")
	utils.GlobalConfig.SetDefault(defaultObjectListPath, "")

}

func (rc *RadarChannels) SetupRunnable(state *utils.State, config *utils.Config) {
	radars := config.GetSettingAsSplit(radarChannelSupportedRadars, ",")
	noRadars := len(radars)

	dataService, ok := state.Get(service.UDPDataServiceName).(*service.UDPData)

	if !ok {
		log.Warn().Msg("Radar Channels not configured due to no UDP data service...")
		return
	}

	rc.InitNoRadars(noRadars)
	rc.AttachTo(dataService)

	for index, radarIP := range radars {
		ip := utils.IP4Builder.FromString(radarIP)
		rc.Radar[index].IPAddress = ip
	}

	rc.Start()
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
		radar.Run(utils.RadarIPOf(index), rc.workflowBuilder)
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
		dataService.IncorrectRadarMetric.Inc(time.Now())
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
