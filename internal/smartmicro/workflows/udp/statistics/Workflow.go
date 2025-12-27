package statistics

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/mixin"
)

type Workflow struct {
	mixin.MixinWorkflow
}

func (w *Workflow) Process(time time.Time, bytes []byte) {
}
