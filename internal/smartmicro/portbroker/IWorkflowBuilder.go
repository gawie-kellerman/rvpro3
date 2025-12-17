package portbroker

import "rvpro3/radarvision.com/utils"

type IWorkflowBuilder interface {
	GetDiagnosticsWorkflow(radarIP utils.IP4) IWorkflow
	GetInstructionWorkflow(radarIP utils.IP4) IWorkflow
	GetPVRWorkflow(radarIP utils.IP4) IWorkflow
	GetTriggerWorkflow(radarIP utils.IP4) IWorkflow
	GetObjectListWorkflow(radarIP utils.IP4) IWorkflow
	GetStatisticsWorkflow(radarIP utils.IP4) IWorkflow
}
