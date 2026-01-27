package eventtrigger

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
)

type Workflow struct {
	Parent       interfaces.IUDPWorkflowParent    `json:"-"`
	Pipeline     *triggerpipeline.TriggerPipeline `json:"-"`
	PipelineItem triggerpipeline.ITriggerPipelineItem
	RadarIP      utils.IP4
	StateName    string
}

func (w *Workflow) Init(parent interfaces.IUDPWorkflowParent) {
	w.Parent = parent
	w.RadarIP = parent.GetRadarIP()
	w.StateName = "Workflow.Event." + w.RadarIP.String()

	// RadarChannels is responsible for registering the global pipeline
	w.Pipeline = utils.GlobalState.
		Get(triggerpipeline.TriggerPipelineStateName).(*triggerpipeline.TriggerPipeline)

	w.PipelineItem = &triggerpipeline.TriggerPipelineOrItem{
		TriggerPipelineItemMixin: triggerpipeline.TriggerPipelineItemMixin{
			Name:    "Base Trigger",
			RadarIP: w.RadarIP,
			Order:   1,
			Status:  triggerpipeline.ChannelStatusCall,
		},
	}

	w.PipelineItem = w.Pipeline.AddItem(w.PipelineItem)

	utils.GlobalState.Set(w.StateName, w)

	w.InitWorkflow()
}

func (w *Workflow) Process(time time.Time, bytes []byte) {
	reader := port.EventTriggerReader{}
	reader.Init(bytes)

	relays := reader.GetRelays()

	if updated := w.PipelineItem.SetTrigger(time, 0, relays); updated {
	}

	// Down-line:
	// 1. Can save the triggers upon update after receive
	// 2. Process the trigger for potential red hold
	// 3. Can save the triggers after processing
}

// InitWorkflow must use the config to determine its execution steps.  These
// individual steps must be setup to be able to execute
func (w *Workflow) InitWorkflow() {
}
