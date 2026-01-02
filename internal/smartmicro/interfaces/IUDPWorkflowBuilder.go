package interfaces

type IUDPWorkflowBuilder interface {
	GetDiagnosticsWorkflow(parent IUDPWorkflowParent) IUDPWorkflow
	GetInstructionWorkflow(parent IUDPWorkflowParent) IUDPWorkflow
	GetPVRWorkflow(parent IUDPWorkflowParent) IUDPWorkflow
	GetTriggerWorkflow(parent IUDPWorkflowParent) IUDPWorkflow
	GetObjectListWorkflow(parent IUDPWorkflowParent) IUDPWorkflow
	GetStatisticsWorkflow(parent IUDPWorkflowParent) IUDPWorkflow
}
