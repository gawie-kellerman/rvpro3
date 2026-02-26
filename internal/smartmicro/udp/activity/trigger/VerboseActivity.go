package trigger

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type VerboseActivity struct {
	MetricName string
	Index      int
	RadarIP4   utils.IP4
	Metrics    VerboseActivityMetrics `json:"-"`
	interfaces.UDPActivityMixin
}

type VerboseActivityMetrics struct {
	UnsupportedVersion *utils.Metric
	utils.MetricsInitMixin
}

func (l *VerboseActivity) Init(workflow interfaces.IUDPWorkflow, index int, metricsName string) {
	l.InitBase(workflow, index, metricsName)
	l.Metrics.InitMetrics(metricsName, &l.Metrics)
}

func (l *VerboseActivity) Process(
	workflow interfaces.IUDPWorkflow,
	_ int,
	tm time.Time,
	bytes []byte,
) {
	th := port.TransportHeaderReader{
		Buffer: bytes,
	}

	ph := port.PortHeaderReader{
		Buffer:      bytes,
		StartOffset: int(th.GetHeaderLength()),
	}

	if ph.GetPortMajorVersion() == 4 && ph.GetPortMinorVersion() == 0 {
		trigger := port.EventTriggerReader{}
		trigger.Init(bytes)

		utils.Print.Fmt(
			"Radar %s Trigger (hi) %016x (lo) %016x\n",
			workflow.GetRadarIP(),
			trigger.GetRelays2(),
			trigger.GetRelays(),
		)
	} else {
		l.Metrics.UnsupportedVersion.IncAt(1, tm)
	}

}
