package broker

import (
	"fmt"
	"reflect"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/utils"
)

type Workflow struct {
	Activities     []interfaces.IUDPActivity
	PortIdentifier uint32
	Metrics        WorkflowMetrics
	Workflows      interfaces.IUDPWorkflows
}

type WorkflowMetrics struct {
	ProcessedCount *utils.Metric
	ProcessedBytes *utils.Metric
	DroppedCount   *utils.Metric
	DroppedBytes   *utils.Metric
	SkippedCount   *utils.Metric
	SkippedBytes   *utils.Metric
	MinDuration    *utils.Metric
	MaxDuration    *utils.Metric
	TotalDuration  *utils.Metric
	utils.MetricsInitMixin
}

func (w *Workflow) GetRadarIP() utils.IP4 {
	return w.Workflows.GetRadarIP()
}

func (w *Workflow) GetPortIdentifier() uint32 {
	return w.PortIdentifier
}

func (w *Workflow) AddActivity(activity interfaces.IUDPActivity) {
	index := len(w.Activities)
	activity.Init(
		w,
		index,
		w.GetActivityName(index, w.GetRadarIP(), activity),
	)
	w.Activities = append(w.Activities, activity)
}

func (w *Workflow) GetActivityName(index int, radarIP utils.IP4, activity interfaces.IUDPActivity) string {
	typeName := reflect.TypeOf(activity).Elem().Name()
	return fmt.Sprintf("Workflow.Activity.[%s].%d.%d.%s", radarIP, w.GetPortIdentifier(), index, typeName)
}

func (w *Workflow) NextActivityId() int {
	return len(w.Activities)
}

func (w *Workflow) Init(workflows interfaces.IUDPWorkflows, portIdentifier uint32) {
	w.Workflows = workflows
	w.PortIdentifier = portIdentifier

	sectionName := fmt.Sprintf(
		"Workflow.Activities.[%s].%d",
		workflows.GetRadarIP(),
		portIdentifier,
	)

	w.Metrics.InitMetrics(sectionName, &w.Metrics)
}

func (w *Workflow) Process(now time.Time, payload []byte) {
	if len(w.Activities) == 0 {
		w.Metrics.SkippedCount.IncAt(1, now)
		w.Metrics.SkippedBytes.IncAt(int64(len(payload)), now)
		return
	}

	w.Metrics.ProcessedCount.IncAt(1, now)
	w.Metrics.ProcessedBytes.IncAt(int64(len(payload)), now)

	startOn := time.Now()
	for index, activity := range w.Activities {
		w.processActivity(activity, index, now, payload)
	}
	endOn := time.Now()

	duration := endOn.Sub(startOn).Milliseconds()
	w.Metrics.MinDuration.SetIfLessAt(duration, endOn)
	w.Metrics.MaxDuration.SetIfMoreAt(duration, endOn)
	w.Metrics.TotalDuration.IncAt(duration, endOn)
}

func (w *Workflow) Drop(now time.Time, payload []byte) {
	w.Metrics.DroppedCount.IncAt(1, now)
	w.Metrics.DroppedBytes.IncAt(int64(len(payload)), now)
}

func (w *Workflow) processActivity(
	activity interfaces.IUDPActivity,
	index int,
	now time.Time,
	payload []byte,
) {
	startOn := time.Now()
	activity.Process(w, index, now, payload)
	endOn := time.Now()

	duration := endOn.Sub(startOn).Milliseconds()
	activity.UpdateMetrics(duration, endOn)
}
