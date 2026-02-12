package udp

import (
	"fmt"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/utils"
)

type WorkflowExecutor struct {
	RadarIP   utils.IP4
	Workflows map[uint32]interfaces.IUDPWorkflow
	Metrics   workflowExecuteMetrics
}

type workflowExecuteMetrics struct {
	MetricsAt      string
	Received       *utils.Metric
	ReceivedBytes  *utils.Metric
	Processed      *utils.Metric
	ProcessedBytes *utils.Metric
	Skipped        *utils.Metric
	SkippedBytes   *utils.Metric
	MinTime        *utils.Metric
	MaxTime        *utils.Metric
	TotalTime      *utils.Metric
}

func (m *workflowExecuteMetrics) Init(radarIP utils.IP4) {
	m.MetricsAt = fmt.Sprintf("Workflows-%s", radarIP.String())
	gm := &utils.GlobalMetrics
	m.Received = gm.U64(m.MetricsAt, "Received")
	m.ReceivedBytes = gm.U64(m.MetricsAt, "Received Bytes")
	m.Processed = gm.U64(m.MetricsAt, "Processed")
	m.ProcessedBytes = gm.U64(m.MetricsAt, "Processed Bytes")
	m.Skipped = gm.U64(m.MetricsAt, "Skipped")
	m.SkippedBytes = gm.U64(m.MetricsAt, "Skipped Bytes")
	m.MinTime = gm.U64(m.MetricsAt, "Min Time")
	m.MaxTime = gm.U64(m.MetricsAt, "Max Time")
}

func (we *WorkflowExecutor) Init(radarIP utils.IP4) {
	we.Metrics.Init(radarIP)
}

func (we *WorkflowExecutor) Execute(now time.Time, portIdentifier uint32, bytes []byte) {
	workflow, ok := we.Workflows[portIdentifier]

	if !ok {
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

	we.Metrics.Processed.Inc(now)
	we.Metrics.ProcessedBytes.AddCount(uint64(len(bytes)), now)

	// Min and Max Time
	duration := endOn.Sub(startOn).Milliseconds()
	we.Metrics.MinTime.ReplaceMinDuration(duration, now)
	we.Metrics.MaxTime.ReplaceMaxDuration(duration, now)
	we.Metrics.TotalTime.AddCount(uint64(duration), now)
}

func (we *WorkflowExecutor) onSkip(now time.Time, bytes []byte) {
	we.Metrics.Skipped.Inc(now)
	we.Metrics.SkippedBytes.AddCount(uint64(len(bytes)), now)
}
