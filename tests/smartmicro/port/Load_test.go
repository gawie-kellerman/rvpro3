package port

import (
	"testing"

	"rvpro3/radarvision.com/internal/smartmicro/port"
)

func TestLoadObjectList1(t *testing.T) {
	objList := port.ObjectList{}
	panicIf(objList.ReadFile(getT45Path("object_list_4.0#3.bin")))
	objList.Th.PrintDetail()
	objList.Ph.PrintDetail()

	ft := port.FlagsType(100)
	ft.IsSkipPayloadCrc()

}

func TestLoadTrigger(t *testing.T) {
	trigger := port.EventTrigger{}
	panicIf(trigger.ReadFile(getT45Path("event_trigger_4.0.bin")))
	trigger.Th.PrintDetail()
	trigger.Ph.PrintDetail()
}

func TestLoadStatistics(t *testing.T) {
	stats := port.Statistics{}
	panicIf(stats.ReadFile(getT45Path("statistics_4.0#2.bin")))
	stats.Th.PrintDetail()
	stats.Ph.PrintDetail()
}

func TestLoadPVR(t *testing.T) {
	pvr := port.PVR{}
	panicIf(pvr.ReadFile(getT45Path("pvr_2.0.bin")))
	pvr.Th.PrintDetail()
	pvr.Ph.PrintDetail()
}
