package interfaces

import (
	"time"
)

type IUDPWorkflow interface {
	GetParent() any
	SetParent(any)
	Process(time.Time, []byte)
}
