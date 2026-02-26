package objectlist

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
	Metrics    LogToConsoleActivityMetrics `json:"-"`
	interfaces.UDPActivityMixin
}

type LogToConsoleActivityMetrics struct {
	UnsupportedVersion *utils.Metric
	utils.MetricsInitMixin
}

func (l *VerboseActivity) Init(
	workflow interfaces.IUDPWorkflow,
	index int,
	fullName string,
) {
	l.InitBase(workflow, index, fullName)
	l.Metrics.InitMetrics(fullName, &l.Metrics)
}

func (l *VerboseActivity) Process(
	workflow interfaces.IUDPWorkflow,
	_ int,
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
