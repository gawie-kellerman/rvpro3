package interfaces

import (
	"time"
)

type NullWorkflow struct {
	MixinWorkflow
}

func (n *NullWorkflow) Process(time time.Time, bytes []byte) {
	// Simply does nothing
}
