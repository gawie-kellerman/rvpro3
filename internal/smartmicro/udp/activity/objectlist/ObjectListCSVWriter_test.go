package objectlist

import (
	"testing"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

func TestObjectListCSVWriter_Write(t *testing.T) {
	wr := ObjectListCSVWriter{}
	wr.SpeedUnit = "mph"
	wr.DistanceUnit = "ft"
	wr.MaxRecords = 5
	wr.Init("/tmp/objlist-%s.%d.csv", "serial", "name", utils.IP4{})

	for n := range 10 {
		utils.Debug.Panic(wr.Write(
			time.Now(),
			n,
			port.OctLongTruck,
			n,
			n,
			180.222,
			190.666,
			12,
			45,
			10.0,
		))
	}
}
