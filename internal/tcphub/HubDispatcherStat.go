package tcphub

type HubDispatcherStat struct {
	ErrCount   uint32
	ErrBytes   uint32
	WriteCount uint32
	WriteBytes uint32
}

func (c *HubDispatcherStat) RegisterWrite(noBytes uint32) {
	c.WriteCount += 1
	c.WriteBytes += noBytes
}

func (c *HubDispatcherStat) RegisterError(noBytes uint32) {
	c.ErrCount += 1
	c.ErrBytes += noBytes
}
