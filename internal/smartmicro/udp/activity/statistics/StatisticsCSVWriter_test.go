package statistics

import (
	"testing"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

func TestStatisticsCSVWriter_Write(t *testing.T) {
	wr := StatisticsCSVWriter{}
	wr.DetectionZones = 3
	wr.IntervalSecs = 30
	wr.SpeedUnit = "mph"

	defer wr.Close()

	wr.Init("/tmp/stats-12-%s.csv", "serial", "name", utils.IP4{})

	for n := range 10 {
		utils.Debug.Panic(wr.Write(time.Now(), 1, port.OctLongTruck, n, 1, 2, 3, 4))
	}
}
