package objectlist

import (
	"fmt"
	"strconv"
	"time"

	"rvpro3/radarvision.com/internal/branding"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/generic"
	"rvpro3/radarvision.com/utils"
)

type ObjectListCSVWriter struct {
	generic.RadarCSVWriterMixin
	FileDate     time.Time
	FileName     string
	FileNo       int
	SpeedUnit    string
	DistanceUnit string
	Records      int
	MaxRecords   int
}

func (o *ObjectListCSVWriter) Init() {
	o.InitWriter(o.onHeader)
	o.CSVFacade.OnFilename = o.onFilenameCallback
}

func (o *ObjectListCSVWriter) onFilenameCallback(provider *utils.CSVRollOverFileWriterProvider) string {
	now := time.Now()

	if o.FileName == "" || !utils.Time.IsSameDay(o.FileDate, now) {
		o.FileName = fmt.Sprintf(o.CSVFacade.PathTemplate, now.Format(utils.FileDateTimeSecond), o.FileNo)
		o.FileDate = now
	}
	return o.FileName

}

func (o *ObjectListCSVWriter) onHeader(provider *utils.CSVRollOverFileWriterProvider, writer *utils.CSVWriter, s string, s2 string) {
	branding.CSVBranding.WriteTitle(writer, "Object List", "3.0.0")
	branding.CSVBranding.WriteSensor(writer, o.SensorSerial, o.SensorName, o.SensorIP)
	branding.CSVBranding.WriteFeaturesNL(writer, "Speed Unit:", o.SpeedUnit, "Distance Unit:", o.DistanceUnit)
	writer.WriteColsNL("TIMESTAMP", "OBJECT ID", "CLASS", "ZONE", "LANE", "HEADING", "SPEED", "LENGTH", "X", "Y")
}

func (o *ObjectListCSVWriter) Write(now time.Time, objectId int, class port.ObjectClassType, zone int, lane int, heading float32, speed float32, length float32, x float32, y float32) error {
	o.Records++
	if o.MaxRecords > 0 {
		if o.Records >= o.MaxRecords {
			o.Records = 0
			o.FileDate = time.Time{}
			o.FileNo += 1
			o.FileName = ""
		}
	}

	w, err := o.CSVFacade.GetWriter()
	if err != nil {
		return err
	}

	w.WriteColsNL(
		now.Format(utils.DisplayDateTimeMS),
		strconv.Itoa(objectId),
		class.String(),
		strconv.Itoa(zone),
		strconv.Itoa(lane),
		strconv.FormatFloat(float64(heading), 'f', 1, 32),
		strconv.FormatFloat(float64(speed), 'f', 1, 32),
		strconv.FormatFloat(float64(length), 'f', 1, 32),
		strconv.FormatFloat(float64(x), 'f', 1, 32),
		strconv.FormatFloat(float64(y), 'f', 1, 32),
	)

	return w.Err
}
