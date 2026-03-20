package generic

import "rvpro3/radarvision.com/utils"

type RadarCSVWriterMixin struct {
	CSVFacade    utils.CSVRollOverFileWriterProvider
	SensorSerial string
	SensorName   string
	SensorIP     string
}

func (t *RadarCSVWriterMixin) InitWriter(
	onHeaderCallback func(*utils.CSVRollOverFileWriterProvider, *utils.CSVWriter, string, string),
) {
	t.CSVFacade.TimeFormat = utils.FileDateTimeSecond
	t.CSVFacade.OnHeader = onHeaderCallback
	t.CSVFacade.OnFilename = t.CSVFacade.OnFileNameCallback
	t.CSVFacade.OnShouldRollover = t.CSVFacade.OnShouldRolloverCallback
}

func (t *RadarCSVWriterMixin) Close() {
	w, err := t.CSVFacade.GetWriter()
	if err != nil {
		return
	}

	_ = w.Flush()
	w.Close()
}

func (t *RadarCSVWriterMixin) Flush() error {
	w, err := t.CSVFacade.GetWriter()
	if err != nil {
		return err
	}
	return w.Flush()
}
