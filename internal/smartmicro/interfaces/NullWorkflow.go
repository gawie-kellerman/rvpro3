package interfaces

import (
	"time"
)

type NullWorkflow struct {
	Parent IUDPWorkflowParent
}

func (w *NullWorkflow) Init(p IUDPWorkflowParent) {
	w.Parent = p
}

func (w *NullWorkflow) Process(time time.Time, bytes []byte) {
	// Simply does nothing
}
