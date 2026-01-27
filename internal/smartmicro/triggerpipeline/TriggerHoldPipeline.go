package triggerpipeline

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

type TriggerHoldPipeline struct {
	TriggerPipelineItemMixin
	DetectChannel     int
	Status            ChannelStatus
	HoldFrom          time.Time
	HoldTo            time.Time
	Initial           ChannelStatus
	Subsequent        ChannelStatus
	DisplayIterations int
}

func (t *TriggerHoldPipeline) Execute(now time.Time, source utils.Uint128, display ITriggerDisplay) utils.Uint128 {
	if now.After(t.HoldTo) {
		t.Triggers.SetBit(t.DetectChannel, false)
	} else {
		if display != nil {
			display.Set(t.DetectChannel, t.GetDisplayStatus())
		}
	}

	return source.Or(t.Triggers)
}

func (t *TriggerHoldPipeline) ReleaseIf(status ChannelStatus) bool {
	if t.Triggers.IsBit(t.DetectChannel) {
		if t.Subsequent == status {
			t.Triggers.SetBit(t.DetectChannel, false)
			return true
		}
	}
	return false
}

func (t *TriggerHoldPipeline) StartHold(
	holdFrom time.Time,
	holdTo time.Time,
	initialStatus ChannelStatus,
	subsequentStatus ChannelStatus,
) {
	t.HoldFrom = holdFrom
	t.HoldTo = holdTo
	t.Initial = initialStatus
	t.Subsequent = subsequentStatus
	t.DisplayIterations = 0
	t.Triggers.SetBit(t.DetectChannel, true)
}

func (t *TriggerHoldPipeline) Is(status ChannelStatus) bool {
	return t.Triggers.IsBit(t.DetectChannel) && t.Subsequent == status
}

func (t *TriggerHoldPipeline) GetDisplayStatus() ChannelStatus {
	t.DisplayIterations++

	if t.DisplayIterations <= 2 {
		return t.Initial
	}
	return t.Subsequent
}
