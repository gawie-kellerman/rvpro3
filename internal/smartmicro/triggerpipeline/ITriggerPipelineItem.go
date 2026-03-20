package triggerpipeline

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

type ITriggerPipelineItem interface {
	GetOrder() int
	GetName() string
	GetRadarIP() utils.IP4
	GetTrigger() utils.Uint128
	SetTrigger(now time.Time, triggerHi uint64, triggerLo uint64) bool
	GetSetOn() time.Time
	GetUpdateOn() time.Time
	SetParent(pipeline *TriggerPipeline)
	GetParent() *TriggerPipeline
	//GetChannelStatus() ChannelStatus
	//SetChannelStatus(channelStatus ChannelStatus)
	Execute(now time.Time, uint128 utils.Uint128, display ITriggerDisplay) utils.Uint128
	SetUpdateOn(tm time.Time)
	// AfterInit is used to initialize "static" variables used in execution.
	//AfterInit()
}
