package interfaces

import (
	"time"
)

type IUDPWorkflow interface {
	SetParent(IUDPWorkflowParent)
	Process(time.Time, []byte)
}
