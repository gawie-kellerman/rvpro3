package triggerpipeline

import (
	"encoding/json"
	"time"

	"rvpro3/radarvision.com/utils"
)

type TriggerPipelineItemMixin struct {
	Order    int
	Name     string
	RadarIP  utils.IP4
	Triggers utils.Uint128
	Status   ChannelStatus
	SetOn    time.Time
	UpdateOn time.Time
}

func (t *TriggerPipelineItemMixin) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Name":       t.Name,
		"Order":      t.Order,
		"RadarIP":    t.RadarIP,
		"Triggers":   t.Triggers.String(),
		"Status":     string(t.Status),
		"StatusName": t.Status.String(),
	})
}

func (t *TriggerPipelineItemMixin) GetOrder() int {
	return t.Order
}

func (t *TriggerPipelineItemMixin) GetName() string {
	return t.Name
}

func (t *TriggerPipelineItemMixin) GetRadarIP() utils.IP4 {
	return t.RadarIP
}

func (t *TriggerPipelineItemMixin) GetTrigger() utils.Uint128 {
	return t.Triggers
}

func (t *TriggerPipelineItemMixin) SetTrigger(now time.Time, hi uint64, lo uint64) bool {
	t.SetOn = now

	if t.Triggers.Hi != hi || t.Triggers.Lo != lo {
		t.UpdateOn = now
		t.Triggers = utils.Uint128{Lo: lo, Hi: hi}
		return true
	}
	return false
}

func (t *TriggerPipelineItemMixin) GetChannelStatus() ChannelStatus {
	return t.Status
}

func (t *TriggerPipelineItemMixin) SetChannelStatus(channelStatus ChannelStatus) {
	t.Status = channelStatus
}

func (t *TriggerPipelineItemMixin) GetSetOn() time.Time {
	return t.SetOn
}

func (t *TriggerPipelineItemMixin) GetUpdateOn() time.Time {
	return t.UpdateOn
}
