package utils

import "testing"

func TestCSVWriterForRollOver_GetWriter(t *testing.T) {
	ro := NewCSVRollOverFileWriterProvider("./hello-%s.csv", FileDateTimeMinute, SampleHeaderCallback)

	csv, err := ro.GetWriter()
	Debug.Panic(err)

	csv.WriteCol("Hello")
	csv.WriteCol("Bob")
	csv.Close()
}

func SampleHeaderCallback(_ *CSVRollOverFileWriterProvider, writer *CSVWriter, ofn string, nfn string) {
	writer.WriteCol(ofn)
	writer.WriteCol(nfn)
	writer.WriteLn()
}
