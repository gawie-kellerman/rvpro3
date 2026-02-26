package interfaces

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

// IUDPActivity is run in the context of a IUDPWorkflow
// It represents a single activity runnable in the context of a
// sensor where the sensor is represented in a data chain of
// IUDPActivity -> IUDPWorkflow -> RadarChannel -> RadarChannels -> UDPData
type IUDPActivity interface {
	GetMetricName() string
	// Init assumes that all struct variables are set already.
	Init(workflow IUDPWorkflow, index int, fullName string)
	Process(time.Time, []byte)
	UpdateMetrics(duration int64, on time.Time)
}

type UDPActivityMixin struct {
	Workflow       IUDPWorkflow
	MetricName     string
	Index          int
	MinDuration    *utils.Metric
	MaxDuration    *utils.Metric
	TotalDuration  *utils.Metric
	ProcessedCount *utils.Metric
	utils.MetricsInitMixin
}

func (u *UDPActivityMixin) InitBase(workflow IUDPWorkflow, index int, metricName string) {
	u.Workflow = workflow
	u.Index = index
	u.MetricName = metricName
	u.InitMetrics(metricName, &u)
}

func (u *UDPActivityMixin) UpdateMetrics(duration int64, now time.Time) {
	u.MinDuration.SetIfLessAt(duration, now)
	u.MaxDuration.SetIfMoreAt(duration, now)
	u.TotalDuration.IncAt(duration, now)
	u.ProcessedCount.IncAt(1, now)
}

func (u *UDPActivityMixin) GetMetricName() string {
	return u.MetricName
}
