package objectlist

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

func (l *LogToConsoleActivity) GetMetricName() string {
	return l.MetricName
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

func (l *LogToConsoleActivity) Process(
	workflow interfaces.IUDPWorkflow,
	i int,
	time time.Time,
	bytes []byte,
) {
	th := port.TransportHeaderReader{
		Buffer: bytes,
	}

	ph := port.PortHeaderReader{
		Buffer:      bytes,
		StartOffset: int(th.GetHeaderLength()),
	}

	if ph.GetPortMajorVersion() == 3 && ph.GetPortMinorVersion() == 0 {
		objList := port.ObjectListReader{}
		objList.Init(bytes)

		utils.Print.Fmt(
			"Radar %s Objects=%d\n",
			workflow.GetRadarIP(),
			objList.GetNofObjects(),
		)
	} else {
		l.Metrics.UnsupportedVersion.IncAt(1, time)
	}
}
