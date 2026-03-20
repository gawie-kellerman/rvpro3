package trigger

import (
	"fmt"
	"strconv"
	"time"

	"rvpro3/radarvision.com/internal/branding"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/generic"
	"rvpro3/radarvision.com/utils"
)

type TriggerCSVWriter struct {
	generic.RadarCSVWriterMixin
}

func (t *TriggerCSVWriter) Init() {
	t.InitWriter(t.onHeader)
}

func (t *TriggerCSVWriter) onHeader(
	_ *utils.CSVRollOverFileWriterProvider,
	writer *utils.CSVWriter,
	_ string,
	_ string,
) {
	branding.CSVBranding.WriteTitle(writer, "Event Trigger", "3.0.0")
	branding.CSVBranding.WriteSensor(writer, t.SensorSerial, t.SensorName, t.SensorIP)
	branding.CSVBranding.WriteFeaturesNL(writer, "")
	writer.WriteColsNL("TIMESTAMP", "TRIGGERED OBJECTS", "TRIGGERED RELAYS", "RELAYS")
}

func (t *TriggerCSVWriter) Write(
	now time.Time,
	objectCount int,
	relayCount int,
	relays uint64,
) error {
	writer, err := t.CSVFacade.GetWriter()
	if err != nil {
		return err
	}

	writer.WriteColsNL(
		now.Format(utils.DisplayDateTimeMS),
		strconv.Itoa(objectCount),
		strconv.Itoa(relayCount),
		fmt.Sprintf("0x%016x", relays),
	)

	return writer.Err
}
