package interfaces

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

type IUDPWorkflow interface {
	GetRadarIP() utils.IP4
	GetPortIdentifier() uint32

	Init(ip utils.IP4, portIdentifier uint32)
	Process(time.Time, []byte)
	Drop(time.Time, []byte)

	AddActivity(activity IUDPActivity)
	NextActivityId() int
}
