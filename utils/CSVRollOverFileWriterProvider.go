package utils

import (
	"fmt"
	"os"
	"path"
	"time"
)

type CSVRollOverFileWriterProvider struct {
	writer           CSVWriter
	PathTemplate     string
	TimeFormat       string
	OnHeader         func(provider *CSVRollOverFileWriterProvider, writer *CSVWriter, oldFilename string, newFilename string)
	OnFilename       func(*CSVRollOverFileWriterProvider) string
	OnShouldRollover func(*CSVRollOverFileWriterProvider) bool
	FileDate         time.Time
}

func NewCSVRollOverFileWriterProvider(
	template string,
	format string,
	onHeaderCallback func(*CSVRollOverFileWriterProvider, *CSVWriter, string, string),
) *CSVRollOverFileWriterProvider {
	res := &CSVRollOverFileWriterProvider{
		PathTemplate: template,
		TimeFormat:   format,
		OnHeader:     onHeaderCallback,
	}
	res.OnFilename = res.OnFileNameCallback
	res.OnShouldRollover = res.OnShouldRolloverCallback
	return res
}

func (c *CSVRollOverFileWriterProvider) GetWriter() (*CSVWriter, error) {
	if c.OnFilename == nil {
		panic("CSVRollOverFileWriterProvider OnFilename is nil")
	}

	if c.OnShouldRollover(c) {
		newFilename := c.OnFilename(c)
		c.writer.Close()
		if err := c.createPathFor(newFilename); err != nil {
			return nil, err
		}
		if err := c.writer.CreateOrOpen(newFilename); err != nil {
			return nil, err
		}

		if c.OnHeader != nil && c.writer.IsNewFile {
			c.OnHeader(c, &c.writer, c.writer.Filename, newFilename)
		}

		c.writer.Filename = newFilename
		c.FileDate = time.Now()

		return &c.writer, nil
	}

	return &c.writer, nil
}

func (c *CSVRollOverFileWriterProvider) OnFileNameCallback(*CSVRollOverFileWriterProvider) string {
	return fmt.Sprintf(c.PathTemplate, Time.Approx().Format(c.TimeFormat))
}

func (c *CSVRollOverFileWriterProvider) OnShouldRolloverCallback(*CSVRollOverFileWriterProvider) bool {
	return !Time.IsSameDay(Time.Approx(), c.FileDate)
}

func (c *CSVRollOverFileWriterProvider) createPathFor(filename string) error {
	return os.MkdirAll(path.Dir(filename), os.ModePerm)
}
