package triggerpipeline

import (
	"encoding/json"
	"fmt"

	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/bit"
)

type TriggerPipelineOrItem struct {
	Order     int
	Name      string
	RadarIP   utils.IP4
	TriggerLo uint64
	TriggerHi uint64
	Status    ChannelStatus
}

func (t *TriggerPipelineOrItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Name":       t.Name,
		"Order":      t.Order,
		"RadarIP":    t.RadarIP,
		"TriggerHi":  fmt.Sprintf("%016x", t.TriggerHi),
		"TriggerLo":  fmt.Sprintf("%016x", t.TriggerLo),
		"Status":     string(t.Status),
		"StatusName": t.Status.String(),
	})
}

func (t *TriggerPipelineOrItem) GetOrder() int {
	return t.Order
}

func (t *TriggerPipelineOrItem) GetName() string {
	return t.Name
}

func (t *TriggerPipelineOrItem) GetRadarIP() utils.IP4 {
	return t.RadarIP
}

func (t *TriggerPipelineOrItem) GetTrigger() (uint64, uint64) {
	return t.TriggerHi, t.TriggerHi
}

func (t *TriggerPipelineOrItem) SetTrigger(hi uint64, lo uint64) {
	t.TriggerHi = hi
	t.TriggerLo = lo
}

func (t *TriggerPipelineOrItem) GetChannelStatus() ChannelStatus {
	return t.Status
}

func (t *TriggerPipelineOrItem) SetChannelStatus(channelStatus ChannelStatus) {
	t.Status = channelStatus
}

func (t *TriggerPipelineOrItem) Execute(hi uint64, lo uint64) (uint64, uint64) {
	return t.TriggerHi | hi, t.TriggerLo | lo
}

func (t *TriggerPipelineOrItem) UpdateDisplay(display ITriggerDisplay) {
	u64 := bit.U64Bits(t.TriggerLo)

	u64.ForEachU64Bit(func(index int, isSet bool) {
		if isSet {
			display.Set(index, t.Status)
		}
	})
}
