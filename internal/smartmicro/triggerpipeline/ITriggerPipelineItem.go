package triggerpipeline

import "rvpro3/radarvision.com/utils"

type ITriggerPipelineItem interface {
	GetOrder() int
	GetName() string
	GetRadarIP() utils.IP4
	GetTrigger() (uint64, uint64)
	SetTrigger(triggerHi uint64, triggerLo uint64)
	GetChannelStatus() ChannelStatus
	SetChannelStatus(channelStatus ChannelStatus)
	Execute(uint64, uint64) (uint64, uint64)
	UpdateDisplay(display ITriggerDisplay)
}
