package workflows

import (
	"fmt"

	"rvpro3/radarvision.com/internal/smartmicro/broker/udp"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/diagnostics"
	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/eventtrigger"
	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/instruction"
	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/objectlist"
	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/pvr"
	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/statistics"
)

type WorkflowBuilder struct {
}

func (w WorkflowBuilder) checkType(parent any) {
	if _, ok := parent.(*udp.RadarChannel); !ok {
		panic(fmt.Sprintf("%v not a RadarChannel", parent))
	}
}

func (w WorkflowBuilder) GetDiagnosticsWorkflow(parent interfaces.IUDPWorkflowParent) interfaces.IUDPWorkflow {
	w.checkType(parent)

	res := new(diagnostics.Workflow)
	res.SetParent(parent)
	return res
}

func (w WorkflowBuilder) GetInstructionWorkflow(parent interfaces.IUDPWorkflowParent) interfaces.IUDPWorkflow {
	w.checkType(parent)

	res := new(instruction.Workflow)
	res.SetParent(parent)
	return res
}

func (w WorkflowBuilder) GetPVRWorkflow(parent interfaces.IUDPWorkflowParent) interfaces.IUDPWorkflow {
	w.checkType(parent)

	res := new(pvr.Workflow)
	res.SetParent(parent)
	return res
}

func (w WorkflowBuilder) GetTriggerWorkflow(parent interfaces.IUDPWorkflowParent) interfaces.IUDPWorkflow {
	w.checkType(parent)

	res := new(eventtrigger.Workflow)
	res.SetParent(parent)
	return res
}

func (w WorkflowBuilder) GetObjectListWorkflow(parent interfaces.IUDPWorkflowParent) interfaces.IUDPWorkflow {
	w.checkType(parent)

	res := new(objectlist.Workflow)
	res.SetParent(parent)
	return res
}

func (w WorkflowBuilder) GetStatisticsWorkflow(parent interfaces.IUDPWorkflowParent) interfaces.IUDPWorkflow {
	w.checkType(parent)

	res := new(statistics.Workflow)
	res.SetParent(parent)
	return res
}
