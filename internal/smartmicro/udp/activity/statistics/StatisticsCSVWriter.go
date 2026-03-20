package statistics

import (
	"strconv"
	"time"

	"rvpro3/radarvision.com/internal/branding"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/generic"
	"rvpro3/radarvision.com/utils"
)

type StatisticsCSVWriter struct {
	generic.RadarCSVWriterMixin
	IntervalSecs   int
	DetectionZones int
	SpeedUnit      string
}

func (t *StatisticsCSVWriter) Init() {
	t.InitWriter(t.onHeader)
}

func (t *StatisticsCSVWriter) onHeader(
	_ *utils.CSVRollOverFileWriterProvider,
	writer *utils.CSVWriter,
	_ string,
	_ string,
) {
	branding.CSVBranding.WriteTitle(writer, "Statistics Data", "3.0.2")
	branding.CSVBranding.WriteSensor(writer, t.SensorSerial, t.SensorName, t.SensorIP)
	branding.CSVBranding.WriteFeaturesNL(
		writer,
		"Interval (secs):", strconv.Itoa(t.IntervalSecs),
		"Detection Zones:", strconv.Itoa(t.DetectionZones),
		"Speed Unit:", t.SpeedUnit,
	)
	writer.WriteColsNL("TIMESTAMP", "ZONE", "CLASS", "VOLUME", "OCCUPANCY", "AVG/SPEED", "HEAD", "GAP")
}

func (t *StatisticsCSVWriter) Write(
	now time.Time,
	zone int,
	class port.ObjectClassType,
	volume int,
	occupancy float32,
	avgSpeed float32,
	head float32,
	gap float32,
) error {
	writer, err := t.CSVFacade.GetWriter()

	if err != nil {
		return err
	}

	writer.WriteColsNL(
		now.Format(utils.DisplayDateTimeMS),
		strconv.Itoa(zone),
		class.String(),
		strconv.Itoa(volume),
		strconv.FormatFloat(float64(occupancy), 'f', 3, 32),
		strconv.FormatFloat(float64(avgSpeed), 'f', 3, 32),
		strconv.FormatFloat(float64(head), 'f', 3, 32),
		strconv.FormatFloat(float64(gap), 'f', 3, 32),
	)
	return writer.Err
}
