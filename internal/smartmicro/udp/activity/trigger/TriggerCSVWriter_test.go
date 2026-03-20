package trigger

import (
	"testing"
	"time"

	"rvpro3/radarvision.com/utils"
)

func TestTriggerCSVWriter_Write(t *testing.T) {
	wr := TriggerCSVWriter{}
	defer wr.Close()
	wr.CSVFacade.PathTemplate = "/tmp/triggers-12-%s.csv"
	wr.SensorIP = "127.0.0.1"
	wr.SensorName = "Sensor"
	wr.SensorSerial = "12345"
	wr.Init()

	for n := range 10 {
		utils.Debug.Panic(wr.Write(time.Now(), 1, 2, uint32(n), 1))
	}
}
