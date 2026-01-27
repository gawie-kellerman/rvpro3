package interfaces

import (
	"time"
)

type IUDPWorkflow interface {
	Init(IUDPWorkflowParent)
	Process(time.Time, []byte)
}
