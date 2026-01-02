package branding

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

type csvBranding struct{}

var CSVBranding csvBranding

func (csvBranding) WriteTitle(writer *utils.CSVWriter, fileType string, fileVersion string) {
	writer.WriteColsNL(fileType, fileVersion)
	writer.WriteColsNL("Radar Vision", "https://radarvision.ai")
	writer.WriteLn("======================================================")
	writer.WriteColsNL("Recording start date:", time.Now().Format(utils.DisplayDateTimeMS))
}

func (csvBranding) WriteFeaturesNL(writer *utils.CSVWriter, data ...string) {
	writer.WriteCol("Features Configured:")
	writer.WriteColsNL(data...)
	writer.WriteLn("------------------------------------------------------")
}

func (csvBranding) WriteFeatures(writer *utils.CSVWriter, data ...string) {
	writer.WriteCol("Features Configured:")
	writer.WriteCols(data...)
}
