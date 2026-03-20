package pvr

import (
	"fmt"
	"strconv"
	"time"

	"rvpro3/radarvision.com/internal/branding"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/generic"
	"rvpro3/radarvision.com/utils"
)

type PVRCSVWriter struct {
	generic.RadarCSVWriterMixin
	FileName     string
	FileDate     time.Time
	FileNo       int
	MaxRecords   int
	Records      int
	SpeedUnit    string
	DistanceUnit string
}

func (p *PVRCSVWriter) Init() {
	p.InitWriter(p.onHeader)
	p.CSVFacade.OnFilename = p.onFilenameCallback
}

func (p *PVRCSVWriter) onHeader(
	_ *utils.CSVRollOverFileWriterProvider,
	writer *utils.CSVWriter,
	_ string,
	_ string,
) {
	branding.CSVBranding.WriteTitle(writer, "Per Vehicle Record", "3.0.0")
	branding.CSVBranding.WriteSensor(writer, p.SensorSerial, p.SensorName, p.SensorIP)
	branding.CSVBranding.WriteFeaturesNL(writer, "Speed Unit:", p.SpeedUnit, "Distance Unit:", p.DistanceUnit)
	writer.WriteColsNL("TIMESTAMP", "OBJECT ID", "CLASS", "ZONE", "HEADING", "SPEED", "LENGTH", "COUNTER")
}

func (p *PVRCSVWriter) Write(
	now time.Time,
	objectId int,
	class port.ObjectClassType,
	zone int,
	heading float32,
	speed float32,
	length float32,
	counter int,
) error {
	p.Records++
	if p.MaxRecords > 0 {
		if p.Records >= p.MaxRecords {
			p.Records = 0
			p.FileDate = time.Time{}
			p.FileNo += 1
			p.FileName = ""
		}
	}

	w, err := p.CSVFacade.GetWriter()
	if err != nil {
		return err
	}

	w.WriteColsNL(
		now.Format(utils.DisplayDateTimeMS),
		strconv.Itoa(objectId),
		class.String(),
		strconv.Itoa(zone),
		strconv.FormatFloat(float64(heading), 'f', 1, 32),
		strconv.FormatFloat(float64(speed), 'f', 1, 32),
		strconv.FormatFloat(float64(length), 'f', 1, 32),
		strconv.Itoa(counter),
	)

	return w.Err
}

func (p *PVRCSVWriter) onFilenameCallback(_ *utils.CSVRollOverFileWriterProvider) string {
	now := time.Now()

	if p.FileName == "" || !utils.Time.IsSameDay(p.FileDate, now) {
		p.FileName = fmt.Sprintf(p.CSVFacade.PathTemplate, now.Format(utils.FileDateTimeSecond), p.FileNo)
		p.FileDate = now
	}
	return p.FileName
}
