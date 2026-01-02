package statistics

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
)

type Workflow struct {
	Parent interfaces.IUDPWorkflowParent
}

func (w *Workflow) SetParent(p interfaces.IUDPWorkflowParent) {
	w.Parent = p
}

func (w *Workflow) Process(time time.Time, bytes []byte) {
}
