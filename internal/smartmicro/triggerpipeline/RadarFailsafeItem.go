package triggerpipeline

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/bit"
)

// RadarFailsafeItem must be registered with the highest order, which will make it
// execute last. It should be setup and executed per radar (RadarChannel).  Its purpose
// Is to simply set/clear flags based on whether the radar is sending message or not.
type RadarFailsafeItem struct {
	TriggerPipelineItemMixin
	SetChannels         utils.Uint128
	ClearChannels       utils.Uint128
	NoRadarActivitySecs int
	MessageProcessedOn  *utils.Metric
}

func (r *RadarFailsafeItem) AfterInit() {
}

func (r *RadarFailsafeItem) Execute(now time.Time, source utils.Uint128, display ITriggerDisplay) utils.Uint128 {
	if r.MessageProcessedOn == nil {
		if !r.InitRadarMetric() {
			panic("radar metric for " + r.RadarIP.String() + " not found")
		}
	}

	processedOn := r.MessageProcessedOn.GetTime()

	if !utils.Time.IsExpired(processedOn, now, time.Duration(r.NoRadarActivitySecs)*time.Second) {
		return source
	}

	lo := bit.U64Bits(r.ClearChannels.Lo)
	lo.ForNotSet(func(index int, isNotSet bool) {
		if isNotSet {
			display.Set(index, ChannelStatusFailSafeOff)
		}
	})

	lo.ForEachBit(func(index int, isSet bool) {
		if isSet {
			display.Set(index, ChannelStatusFailSafeOn)
		}
	})

	res := source.And(r.ClearChannels)
	res = res.Or(r.SetChannels)
	return res
}

func (r *RadarFailsafeItem) InitRadarMetric() bool {
	sectionName := interfaces.MetricName.GetUDPRadarMetric(r.RadarIP)
	r.MessageProcessedOn = utils.GlobalMetrics.
		Section(sectionName).
		Get(interfaces.MessageProcessedOnMetricName)
	return r.MessageProcessedOn != nil
}
