package ping

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus-community/pro-bing"
	"rvpro3/radarvision.com/utils"
)

const DeviceTypeCamera = "Camera"
const DeviceTypeRadar = "Radar"

type PingStat struct {
	IP         utils.IP4
	DriverStat probing.Statistics
	LastError  error
	LastPing   time.Time
	Metrics    PingStatMetrics
	DeviceType string
}

type PingStatMetrics struct {
	PingOkCount   *utils.Metric
	PingFailCount *utils.Metric
	utils.MetricsInitMixin
}

func (p *PingStat) Init(ip4 utils.IP4, deviceType string) {
	metricName := fmt.Sprintf("Ping.%s", ip4)
	p.Metrics.InitMetrics(metricName, &p.Metrics)
	p.DriverStat.Addr = ip4.ToIPString()
	p.DeviceType = deviceType
	p.IP = ip4
}

func (p *PingStat) SetFail(err error) {
	p.LastError = err
	p.LastPing = utils.Time.Approx()
}

func (p *PingStat) SetSuccess(statistics *probing.Statistics) {
	p.LastError = nil
	p.LastPing = time.Now()
	p.DriverStat = *statistics
}

func (p *PingStat) AddStatusTo(bld *strings.Builder) {
	src := p.DriverStat

	if src.PacketsRecv == src.PacketsSent && src.PacketsSent > 0 {
		bld.WriteString(fmt.Sprintf("[%d", int(p.IP.GetHost())))
	} else if src.PacketsRecv == 0 {
		bld.WriteString("[ ")
	} else {
		bld.WriteString("[!")
	}

	if p.LastError != nil {
		bld.WriteRune('*')
	}

	bld.WriteRune(']')
}
