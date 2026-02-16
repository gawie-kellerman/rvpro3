package trigger

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type LogToConsoleActivity struct {
	MetricName string
	Index      int
	RadarIP4   utils.IP4
	Metrics    LogToConsoleActivityMetrics `json:"-"`
	Mixin      interfaces.UDPActivityMixin
}

type LogToConsoleActivityMetrics struct {
	UnsupportedVersion *utils.Metric
	utils.MetricsInitMixin
}

func (l *LogToConsoleActivity) Init(index int, radarIP utils.IP4, fullName string) {
	l.Index = index
	l.RadarIP4 = radarIP
	l.MetricName = fullName

	l.Mixin.InitMetrics(fullName, &l.Mixin)
	l.Metrics.InitMetrics(fullName, &l.Metrics)
}

func (l *LogToConsoleActivity) SetDuration(duration int64, on time.Time) {
	l.Mixin.SetDuration(duration, on)
}

func (l *LogToConsoleActivity) GetMetricName() string {
	return l.MetricName
}

func (l *LogToConsoleActivity) Process(
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
