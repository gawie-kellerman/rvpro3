package eventtrigger

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/bit"
)

// CallHoldActivity
// Note that both DetectChannel and RedPhaseIndex is 1 based by configured and
// 0 based for system purposes, meaning you have to deduct 1 for system usage and add
// 1 for configuration usage
type CallHoldActivity struct {
	DetectChannel int
	RedPhaseIndex int
	PreviousHold  callHold
	CurrentHold   callHold

	TriggerHoldSecs int
	RedHoldSecs     int

	PipelineItem triggerpipeline.TriggerHoldPipeline
	phaseState   *interfaces.PhaseState
}

type callHold struct {
	IsRed     bool
	IsTrigger bool
}

func (c *CallHoldActivity) Init(parent interfaces.IUDPWorkflowParent) {
	c.phaseState = utils.GlobalState.Get(interfaces.PhaseStateName).(*interfaces.PhaseState)

	c.PipelineItem.Name = fmt.Sprintf(
		"Call Hold: %s Detect %d Phase %d",
		parent.GetRadarIP(),
		c.DetectChannel,
		c.DetectChannel,
	)

	c.PipelineItem.DetectChannel = c.DetectChannel
	c.PipelineItem.Order = 5
	c.PipelineItem.RadarIP = parent.GetRadarIP()

	pipeline := utils.GlobalState.Get(triggerpipeline.TriggerPipelineStateName).(*triggerpipeline.TriggerPipeline)
	pipeline.AddItem(&c.PipelineItem)
}

func (c *CallHoldActivity) Process(time time.Time, readerAny any) {
	reader, ok := readerAny.(*port.EventTriggerReader)

	if !ok {
		panic(errors.New("reader is not a EventTriggerReader"))
	}

	c.PreviousHold.IsRed = c.CurrentHold.IsRed
	c.PreviousHold.IsTrigger = c.CurrentHold.IsTrigger

	red, _, _ := c.phaseState.GetRYG()

	triggers := reader.GetRelays()
	c.CurrentHold.IsRed = bit.IsSet(red, c.RedPhaseIndex)
	c.CurrentHold.IsTrigger = bit.IsSet(triggers, c.TriggerHoldSecs)

	/*
		When you receive a trigger event, then there is no need to hold
		the trigger anymore
	*/
	if c.CurrentHold.IsTrigger {
		c.PipelineItem.ReleaseIf(triggerpipeline.ChannelStatusRedExtend)
	}

	if c.PreviousHold.IsRed {
		if c.CurrentHold.IsRed {
			c.doRedToRed(time, reader)
		} else {
			c.doRedToGreen(time, reader)
		}
	} else {
		if c.CurrentHold.IsRed {
			c.doGreenToRed(time)
		} else {
			// Nothing to do here
		}
	}
}

func (c *CallHoldActivity) doRedToRed(now time.Time, reader *port.EventTriggerReader) {
	c.PipelineItem.ReleaseIf(triggerpipeline.ChannelStatusRedHold)

	switch c.PreviousHold.IsTrigger {
	// There was a previous trigger
	case true:
		switch c.CurrentHold.IsTrigger {

		case true:
			// Red to Red, Trigger to trigger, already release at line 62ish
			break

		case false:
			// Red to Red, Trigger to no trigger, start to hold the trigger
			c.PipelineItem.StartHold(
				now,
				now.Add(time.Duration(c.TriggerHoldSecs)*time.Second),
				triggerpipeline.ChannelStatusRedExtend,
				triggerpipeline.ChannelStatusRedExtend,
			)
		}

	// There is no previous trigger
	case false:
		// Nothing to do as the normal event propagation takes care of the state
		break
	}
}

func (c *CallHoldActivity) doRedToGreen(now time.Time, reader *port.EventTriggerReader) {
	if c.PipelineItem.Is(triggerpipeline.ChannelStatusRedExtend) {
		// If holding a trigger while robot turns from Red to Green
		// then extend using a Red Hold.
		// By using the same detectChannel the Trigger Hold is automatically
		// overwritten with a Red Hold

		c.PipelineItem.StartHold(
			now,
			now.Add(time.Duration(c.RedHoldSecs)*time.Second),
			triggerpipeline.ChannelStatusRedHold,
			triggerpipeline.ChannelStatusRedHold,
		)
	}
}

func (c *CallHoldActivity) doGreenToRed(time.Time) {
	c.PipelineItem.ReleaseIf(triggerpipeline.ChannelStatusRedHold)
}
