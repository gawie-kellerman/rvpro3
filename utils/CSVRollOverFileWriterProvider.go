package utils

import (
	"fmt"
	"time"
)

type CSVRollOverFileWriterProvider struct {
	writer     CSVWriter
	Template   string
	Format     string
	OnHeader   func(provider *CSVRollOverFileWriterProvider, writer *CSVWriter, oldFilename string, newFilename string)
	OnFilename func(*CSVRollOverFileWriterProvider) string
}

func NewCSVRollOverFileWriterProvider(
	template string,
	format string,
	onHeaderCallback func(*CSVRollOverFileWriterProvider, *CSVWriter, string, string),
) *CSVRollOverFileWriterProvider {
	res := &CSVRollOverFileWriterProvider{
		Template: template,
		Format:   format,
		OnHeader: onHeaderCallback,
	}
	res.OnFilename = res.OnFileNameCallback
	return res
}

func (c *CSVRollOverFileWriterProvider) GetWriter() (*CSVWriter, error) {
	if c.OnFilename == nil {
		panic("CSVRollOverFileWriterProvider OnFilename is nil")
	}

	newFilename := c.OnFilename(c)
	oldFilename := c.writer.Filename

	if newFilename != c.writer.Filename {
		c.writer.Close()
		err := c.writer.CreateOrOpen(newFilename)

		if err != nil {
			return nil, err
		}

		if c.OnHeader != nil && c.writer.IsNewFile {
			c.OnHeader(c, &c.writer, oldFilename, newFilename)
		}

		return &c.writer, err
	}

	return &c.writer, nil
}

func (c *CSVRollOverFileWriterProvider) OnFileNameCallback(*CSVRollOverFileWriterProvider) string {
	return fmt.Sprintf(c.Template, time.Now().Format(c.Format))
}
