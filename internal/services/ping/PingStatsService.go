package ping

import (
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"rvpro3/radarvision.com/utils"
)

const PingStatsServiceName = "Ping.Stats.Service"
const cameraIPs = "192.168.11.22:443;192.168.11.23:443;192.168.11.24:443;192.168.11.25:443"
const radarIPs = "192.168.11.12:55555;192.168.11.13:55555;192.168.11.14:55555;192.168.11.15:55555"

type PingStatsService struct {
	Terminate          bool
	Terminated         bool
	IsEnabled          bool
	IsReady            bool
	Stats              PingStats
	pinger             *probing.Pinger
	initialCooldown    utils.Milliseconds
	subsequentCooldown utils.Milliseconds
	cooldown           utils.Milliseconds
	pingTimeout        time.Duration
	pingCount          int
}

func (p *PingStatsService) InitFromSettings(settings *utils.Settings) {
	p.Stats.Init()

	p.IsEnabled = settings.Basic.GetBool("ping.enabled", true)
	p.setupPings(DeviceTypeRadar, settings.Basic.GetArray("ping.radar.ips", radarIPs))
	p.setupPings(DeviceTypeCamera, settings.Basic.GetArray("ping.camera.ips", cameraIPs))
	p.initialCooldown = settings.Basic.GetMilliseconds("ping.initial.cooldown", 50)
	p.subsequentCooldown = settings.Basic.GetMilliseconds("ping.subsequent.cooldown", 2000)
	p.pingTimeout = time.Duration(settings.Basic.GetMilliseconds("ping.timeout", 1000))
	p.pingCount = settings.Basic.GetInt("ping.count", 1)

	p.cooldown = p.initialCooldown
	p.IsReady = true
}

func (p *PingStatsService) Sprint() {
	p.cooldown = p.initialCooldown
}

func (p *PingStatsService) Start(state *utils.State, settings *utils.Settings) {
	if p.IsEnabled {
		go p.run()
	}
}

func (p *PingStatsService) GetServiceName() string {
	return PingStatsServiceName
}

func (p *PingStatsService) run() {
	for !p.Terminated {
		if p.IsReady {
			p.pingNext()
		}
		p.cooldown.Sleep()
	}

	p.Terminated = true
}

func (p *PingStatsService) pingNext() {
	p.IsReady = false

	var err error
	current, isWrapped := p.Stats.GetNext()

	if isWrapped {
		p.cooldown = p.subsequentCooldown
	}

	if current == nil {
		return
	}

	p.pinger, err = probing.NewPinger(current.DriverStat.Addr)

	if err != nil {
		current.LastError = err
		current.LastPing = utils.Time.Approx()
		return
	}

	p.pinger.Timeout = p.pingTimeout
	p.pinger.Count = p.pingCount
	p.pinger.Size = 32

	err = p.pinger.SetAddr(current.DriverStat.Addr)
	p.pinger.SetPrivileged(true)
	p.pinger.Interval = 1 * time.Second

	if err != nil {
		current.LastError = err
		current.LastPing = utils.Time.Approx()
		return
	}

	p.pinger.OnFinish = func(statistics *probing.Statistics) {
		current.SetSuccess(statistics)
		p.IsReady = true
	}

	err = p.pinger.Run()

	if err != nil {
		current.SetFail(err)
		p.IsReady = true
	}
}

func (p *PingStatsService) setupPings(deviceType string, ipList []string) {
	for _, ip := range ipList {
		ip4 := utils.IP4Builder.FromString(ip)
		p.Stats.Add(ip4, deviceType)
	}
}
