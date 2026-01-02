package eventtrigger

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/bit"
)

type Workflow struct {
	Parent       interfaces.IUDPWorkflowParent
	Pipeline     *triggerpipeline.TriggerPipeline
	PipelineItem triggerpipeline.ITriggerPipelineItem
}

func (w *Workflow) SetParent(parent interfaces.IUDPWorkflowParent) {
	w.Parent = parent
	w.Pipeline = utils.GlobalState.
		Get(triggerpipeline.TriggerPipelineStateName).(*triggerpipeline.TriggerPipeline)

	w.PipelineItem = w.Pipeline.AddItem(&triggerpipeline.TriggerPipelineOrItem{
		Name:    "Base Trigger",
		RadarIP: parent.GetRadarIP(),
		Order:   1,
		Status:  triggerpipeline.ChannelStatusCall,
	})
}

func (w *Workflow) Process(time time.Time, bytes []byte) {
	reader := port.EventTriggerReader{}
	reader.Init(bytes)

	lo := reader.GetRelays1() + 10
	hi := reader.GetRelays2()
	relays := bit.CombineU32(hi, lo)

	w.PipelineItem.SetTrigger(0, relays)

	// Down-line:
	// 1. Can save the triggers upon update after receive
	// 2. Process the trigger for potential red hold
	// 3. Can save the triggers after processing
}
