package interfaces

import "rvpro3/radarvision.com/utils"

type IUDPWorkflows interface {
	GetRadarIP() utils.IP4
	HasWorkflow(uint32) bool
	GetWorkflow(uint32) IUDPWorkflow
}
