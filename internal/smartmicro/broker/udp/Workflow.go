package udp

import (
	"fmt"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/utils"
)

type Workflow struct {
	RadarIP        utils.IP4
	Activities     []interfaces.IUDPActivity
	PortIdentifier uint32
	Metrics        workflowMetrics
}

func (w *Workflow) GetRadarIP() utils.IP4 {
	return w.RadarIP
}

func (w *Workflow) GetPortIdentifier() uint32 {
	return w.PortIdentifier
}

type workflowMetrics struct {
	MetricAt       string
	Processed      *utils.Metric
	ProcessedBytes *utils.Metric
	Dropped        *utils.Metric
	DroppedBytes   *utils.Metric
	Skipped        *utils.Metric
	SkippedBytes   *utils.Metric
	MinTime        *utils.Metric
	MaxTime        *utils.Metric
}

func (w *workflowMetrics) Init(ip utils.IP4, portIdentifier uint32) {
	w.MetricAt = fmt.Sprintf("Radar %s Identifier %d", ip, portIdentifier)
	gm := &utils.GlobalMetrics

	w.Processed = gm.U64(w.MetricAt, "Processed")
	w.ProcessedBytes = gm.U64(w.MetricAt, "Processed Bytes")
	w.Skipped = gm.U64(w.MetricAt, "Skipped")
	w.SkippedBytes = gm.U64(w.MetricAt, "Skipped Bytes")
	w.Dropped = gm.U64(w.MetricAt, "Dropped")
	w.DroppedBytes = gm.U64(w.MetricAt, "Dropped Bytes")
	w.MinTime = gm.U64(w.MetricAt, "MinTime")
	w.MaxTime = gm.U64(w.MetricAt, "MaxTime")
}

func (w *Workflow) Init(ip utils.IP4, portIdentifier uint32) {
	w.RadarIP = ip
	w.PortIdentifier = portIdentifier
	w.Metrics.Init(ip, portIdentifier)
}

func (w *Workflow) Process(now time.Time, payload []byte) {
	if len(w.Activities) == 0 {
		w.Metrics.Skipped.Inc(now)
		w.Metrics.SkippedBytes.Add(len(payload), now)
		return
	}

	w.Metrics.Processed.Inc(now)
	w.Metrics.ProcessedBytes.Add(len(payload), now)

	startOn := time.Now()
	for index, activity := range w.Activities {
		activity.Process(w, index, now, payload)
	}
	endOn := time.Now()

	duration := endOn.Sub(startOn).Milliseconds()
	w.Metrics.MinTime.ReplaceMinDuration(duration, endOn)
	w.Metrics.MaxTime.ReplaceMaxDuration(duration, now)
}

func (w *Workflow) Drop(now time.Time, payload []byte) {
	w.Metrics.Dropped.Inc(now)
	w.Metrics.DroppedBytes.Add(len(payload), now)
}
