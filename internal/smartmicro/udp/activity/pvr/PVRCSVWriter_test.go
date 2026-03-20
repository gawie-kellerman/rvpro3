package pvr

import (
	"testing"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

func TestPVRCSVWriter_Write(t *testing.T) {
	wr := PVRCSVWriter{}
	wr.SpeedUnit = "mph"
	wr.DistanceUnit = "ft"
	wr.MaxRecords = 5
	wr.Init("/tmp/pvr-12-%s-%d.csv", "serial", "name", utils.IP4{})

	for n := range 20 {
		utils.Debug.Panic(wr.Write(time.Now(), n, port.OctLongTruck, n, 1.234, 27.554, 3.3, n))
		time.Sleep(29 * time.Millisecond)
	}
}
