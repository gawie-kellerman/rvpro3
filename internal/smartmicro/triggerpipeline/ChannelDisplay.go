package triggerpipeline

type ChannelDisplay struct {
	Status   [124]ChannelStatus
	Statuses int
}

func (c *ChannelDisplay) Clear(length int, status ChannelStatus) {
	c.Statuses = length
	for i := range c.Statuses {
		c.Status[i] = status
	}
}

func (c *ChannelDisplay) String() string {
	return string(c.Status[:c.Statuses])
}

func (c *ChannelDisplay) Set(index int, status ChannelStatus) {
	c.Status[index] = status
}

func (c *ChannelDisplay) Get(index int) ChannelStatus {
	return c.Status[index]
}

func (c *ChannelDisplay) ChangeOtherThan(otherThan ChannelStatus, changeTo ChannelStatus) {
	for i := range c.Statuses {
		if c.Status[i] != otherThan {
			c.Status[i] = changeTo
		}
	}
}
