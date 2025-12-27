package common

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/mixin"
)

type NullWorkflow struct {
	mixin.MixinWorkflow
}

func (n *NullWorkflow) Process(time time.Time, bytes []byte) {
	// Simply does nothing
}
