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
	Init(index int, radarIP utils.IP4, fullName string)
	Process(IUDPWorkflow, int, time.Time, []byte)
	SetDuration(duration int64, on time.Time)
}

type UDPActivityMixin struct {
	MinDuration    *utils.Metric
	MaxDuration    *utils.Metric
	TotalDuration  *utils.Metric
	ProcessedCount *utils.Metric
	utils.MetricsInitMixin
}

func (u *UDPActivityMixin) SetDuration(duration int64, now time.Time) {
	u.MinDuration.SetIfLessAt(duration, now)
	u.MaxDuration.SetIfMoreAt(duration, now)
	u.TotalDuration.IncAt(duration, now)
	u.ProcessedCount.IncAt(1, now)
}
