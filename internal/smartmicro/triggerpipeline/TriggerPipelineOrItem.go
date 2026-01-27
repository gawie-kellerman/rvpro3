package triggerpipeline

import (
	"time"

	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/bit"
)

type TriggerPipelineOrItem struct {
	TriggerPipelineItemMixin
}

func (t TriggerPipelineOrItem) Execute(now time.Time, source utils.Uint128, display ITriggerDisplay) utils.Uint128 {
	if display != nil {
		u64 := bit.U64Bits(t.Triggers.Lo)

		u64.ForEachBit(func(index int, isSet bool) {
			if isSet {
				display.Set(index, t.Status)
			}
		})
	}

	return source.Or(t.Triggers)
}
