package interfaces

import "time"

type IUDPWorkflowActivity interface {
	Init(IUDPWorkflowParent)
	Process(time.Time, any)
}
