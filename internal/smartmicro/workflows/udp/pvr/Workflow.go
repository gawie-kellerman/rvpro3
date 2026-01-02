package pvr

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
)

type Workflow struct {
	interfaces.MixinWorkflow
}

func (w *Workflow) Process(time time.Time, bytes []byte) {
}
