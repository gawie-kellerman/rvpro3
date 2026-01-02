package utils

import "testing"

func TestCsvWriter_CreateOrOpen(t *testing.T) {
	csv, err := CsvFile.CreateOrOpen("test.csv")
	Debug.Panic(err)
	defer csv.Close()

	if csv.IsNewFile {
		csv.WriteCols("Header 1", "Header 2", "Header 3", "Header 4")
	}
	csv.WriteF64(100.12345, 2)
	Debug.Panic(csv.Err)
}

func BenchmarkCsvWriter_CreateOrOpenAlloc(b *testing.B) {
	b.ReportAllocs()
	csv, err := CsvFile.CreateOrOpen("test.csv")
	Debug.Panic(err)
	defer csv.Close()

	if csv.IsNewFile {
		csv.WriteCols("Header 1", "Header 2", "Header 3", "Header 4")
	}
	Debug.Panic(csv.Err)
}
