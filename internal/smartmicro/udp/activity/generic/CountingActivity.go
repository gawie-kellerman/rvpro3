package generic

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type CountingActivity struct {
	StartTime  time.Time
	OpenTime   time.Time
	CloseTime  time.Time
	CycleCount int
	CycleBytes int
	TotalCount int
	TotalBytes int
	interfaces.UDPActivityMixin
}

func (c *CountingActivity) Init(
	workflow interfaces.IUDPWorkflow,
	index int,
	metricsName string,
) {
	c.InitBase(workflow, index, metricsName)
	c.OpenTime = time.Now()
}

func (c *CountingActivity) Process(
	_ interfaces.IUDPWorkflow,
	_ int,
	time time.Time,
	bytes []byte,
) {
	c.CycleCount++
	c.CycleBytes += len(bytes)

	c.TotalCount++
	c.TotalBytes += len(bytes)

	c.CloseTime = time

	diff := c.CloseTime.Sub(c.OpenTime)
	if diff.Seconds() > 10 {
		c.dumpOutput(diff)

		c.OpenTime = c.CloseTime
		c.CycleCount = 0
		c.CycleBytes = 0
	}
}

func (c *CountingActivity) dumpOutput(duration time.Duration) {
	pid := port.PortIdentifier(c.ActivityType)

	millis := duration.Milliseconds()
	if millis == 0 {
		utils.Print.Ln("Measurement too small")
	} else {
		cycleSeconds := float64(millis) / float64(1000)
		totalSeconds := float64(c.CloseTime.Sub(c.StartTime).Milliseconds()) / float64(1000)

		utils.Print.Fmt("Radar: %s-%s, ", c.RadarIP4, pid)
		utils.Print.Fmt("Cycle[Dur: %d, Cnt: %d, Avg/Sec: %.2f, Bytes: %d, Bytes/Sec: %d], ",
			cycleSeconds,
			c.CycleCount,
			float64(c.CycleCount)/cycleSeconds,
			c.CycleBytes,
			float64(c.CycleBytes)/cycleSeconds)

		utils.Print.Fmt("Total[Dur: %d, Cnt: %d, Avg/Sec: %.2f, Bytes: %d, Bytes/Sec: %d]\n",
			totalSeconds,
			c.TotalCount,
			float64(c.TotalCount)/cycleSeconds,
			c.TotalBytes,
			float64(c.TotalBytes)/cycleSeconds)
	}
}
