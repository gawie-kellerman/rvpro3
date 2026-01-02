package utils

import (
	"io"
	"os"
	"strconv"
)

var comma = []byte{','}
var newLine = []byte{'\n'}

var CsvFile csvWriterHelper

type csvWriterHelper struct {
}

func (csvWriterHelper) CreateOrOpen(filename string) (res *CSVWriter, err error) {
	res = &CSVWriter{}
	err = res.CreateOrOpen(filename)

	if err != nil {
		return nil, err
	}

	return res, nil
}

type CSVWriter struct {
	Filename   string
	IsNewFile  bool
	Err        error
	writer     io.Writer
	lineWrites int
}

func (c *CSVWriter) CreateOrOpen(filename string) error {
	c.Filename = filename
	c.IsNewFile, c.Err = File.Exists(filename)
	c.IsNewFile = !c.IsNewFile

	if c.Err != nil {
		return c.Err
	}

	if c.IsNewFile {
		// Create the writer
		c.writer, c.Err = os.Create(filename)
	} else {
		// Open the writer.  An opened writer is assumed to end with a newline
		// implying that lineWrites is false
		c.writer, c.Err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	}

	c.lineWrites = 0
	return c.Err
}

func (c *CSVWriter) Close() {
	if c.writer != nil {
		if c.lineWrites > 0 {
			c.WriteNL()
		}

		_ = c.Flush()
		if closer, ok := c.writer.(io.Closer); ok {
			_ = closer.Close()
		}
		c.writer = nil
		c.Filename = ""
	}
}

func (c *CSVWriter) WriteLn(value string) {
	c.WriteCol(value)
	c.WriteNL()
}

func (c *CSVWriter) WriteNL() {
	_, c.Err = c.writer.Write(newLine)
	c.lineWrites = 0
}

func (c *CSVWriter) WriteCol(value string) {
	if c.lineWrites > 0 {
		_, c.Err = c.writer.Write(comma)
	}
	_, c.Err = c.writer.Write([]byte(value))
	c.lineWrites += 1
}

func (c *CSVWriter) WriteColNL(value string) {
	c.WriteCol(value)
	c.WriteNL()
}

func (c *CSVWriter) WriteCols(values ...string) {
	for _, value := range values {
		c.WriteCol(value)
	}
}

func (c *CSVWriter) WriteColsNL(values ...string) {
	for _, value := range values {
		c.WriteCol(value)
	}

	c.WriteNL()
}

func (c *CSVWriter) WriteF64(value float64, precision int) {
	res := strconv.FormatFloat(value, 'f', precision, 64)
	c.WriteCol(res)
}

func (c *CSVWriter) WriteF32(value float64, precision int) {
	res := strconv.FormatFloat(value, 'f', precision, 32)
	c.WriteCol(res)
}

func (c *CSVWriter) WriteInt(value int) {
	c.WriteCol(strconv.Itoa(value))
}

func (c *CSVWriter) Flush() error {
	if fl, ok := c.writer.(*os.File); ok {
		return fl.Sync()
	}

	return nil
}
