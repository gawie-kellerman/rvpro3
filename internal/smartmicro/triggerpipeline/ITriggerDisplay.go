package triggerpipeline

type ITriggerDisplay interface {
	Set(index int, status ChannelStatus)
}
