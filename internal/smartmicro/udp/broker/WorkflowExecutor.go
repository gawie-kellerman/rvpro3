package broker

import (
	"fmt"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/utils"
)

type WorkflowExecutor struct {
	RadarIP   utils.IP4
	Workflows map[uint32]interfaces.IUDPWorkflow
	Metrics   WorkflowExecutorMetrics
}

type WorkflowExecutorMetrics struct {
	ProcessedCount       *utils.Metric
	ProcessedBytes       *utils.Metric
	ProcessedDuration    *utils.Metric
	ProcessedMinDuration *utils.Metric
	ProcessedMaxDuration *utils.Metric
	SkippedCount         *utils.Metric
	SkippedBytes         *utils.Metric
	utils.MetricsInitMixin
}

func (we *WorkflowExecutor) Workflow(portIdentifier uint32) interfaces.IUDPWorkflow {
	if we.Workflows == nil {
		we.Workflows = make(map[uint32]interfaces.IUDPWorkflow)
	}

	res, ok := we.Workflows[portIdentifier]
	if !ok {
		res = &Workflow{}
		res.Init(we.RadarIP, portIdentifier)
		we.Workflows[portIdentifier] = res
	}
	return res
}

func (we *WorkflowExecutor) Init(radarIP utils.IP4) {
	we.RadarIP = radarIP
	we.Metrics.InitMetrics(fmt.Sprintf("Workflow.Executor.[%s]", radarIP), &we.Metrics)
}

func (we *WorkflowExecutor) Execute(
	now time.Time,
	portIdentifier uint32,
	bytes []byte,
) {
	workflow, ok := we.Workflows[portIdentifier]

	if ok {
		we.onProcess(now, workflow, bytes)
	} else {
		we.onSkip(now, bytes)
	}
}

func (we *WorkflowExecutor) Drop(now time.Time, portIdentifier uint32, bytes []byte) {
	workflow, ok := we.Workflows[portIdentifier]

	if !ok {
		workflow = &Workflow{}
		workflow.Init(we.RadarIP, portIdentifier)
		we.Workflows[portIdentifier] = workflow
	}

	workflow.Drop(now, bytes)
}

func (we *WorkflowExecutor) onProcess(now time.Time, workflow interfaces.IUDPWorkflow, bytes []byte) {
	startOn := time.Now()
	workflow.Process(now, bytes)
	endOn := time.Now()

	we.Metrics.ProcessedCount.IncAt(1, now)
	we.Metrics.ProcessedBytes.IncAt(int64(len(bytes)), now)

	// Min and Max Time
	duration := endOn.Sub(startOn).Milliseconds()
	we.Metrics.ProcessedMinDuration.SetIfLessAt(duration, now)
	we.Metrics.ProcessedMaxDuration.SetIfMoreAt(duration, now)
	we.Metrics.ProcessedDuration.IncAt(duration, now)
}

func (we *WorkflowExecutor) onSkip(now time.Time, bytes []byte) {
	we.Metrics.SkippedCount.IncAt(1, now)
	we.Metrics.SkippedBytes.IncAt(int64(len(bytes)), now)
}
