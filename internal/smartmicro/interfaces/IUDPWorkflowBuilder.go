package interfaces

type IUDPWorkflowBuilder interface {
	GetDiagnosticsWorkflow(parent any) IUDPWorkflow
	GetInstructionWorkflow(parent any) IUDPWorkflow
	GetPVRWorkflow(parent any) IUDPWorkflow
	GetTriggerWorkflow(parent any) IUDPWorkflow
	GetObjectListWorkflow(parent any) IUDPWorkflow
	GetStatisticsWorkflow(parent any) IUDPWorkflow
}
