package gpio

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Direction int

const exportPath = "/sys/class/gpio/export"
const unexportPath = "/sys/class/gpio/unexport"
const directionPath = "/sys/class/gpio/gpio%d/direction"
const valuePath = "/sys/class/gpio/gpio%d/value"

var writeVal = [2]string{"0", "1"}

const (
	DirectionIn Direction = iota
	DirectionOut
)

type fileSystem struct{}

var FileSystem fileSystem

func (fileSystem) UnexportPort(port int) (err error) {
	var file *os.File
	if file, err = os.OpenFile(unexportPath, os.O_WRONLY, 755); err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	portStr := strconv.Itoa(port)

	_, err = file.WriteString(portStr)
	return err
}

func (fileSystem) ExportPort(port int) (err error) {
	var file *os.File
	if file, err = os.OpenFile(exportPath, os.O_WRONLY, 755); err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	portStr := strconv.Itoa(port)

	_, err = file.WriteString(portStr)
	return err
}

func (fileSystem) WriteDirection(port int, direction Direction) (err error) {
	var value string
	filename := fmt.Sprintf(directionPath, port)

	if direction == DirectionIn {
		value = "in"
	} else {
		value = "out"
	}

	var file *os.File
	if file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 755); err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.WriteString(value)
	return err
}

func (fileSystem) Open(port int, direction Direction) (file *os.File, err error) {
	filename := fmt.Sprintf(valuePath, port)
	switch direction {
	case DirectionIn:
		return os.OpenFile(filename, os.O_RDONLY, 755)

	case DirectionOut:
		return os.OpenFile(filename, os.O_WRONLY, 755)
	}

	return nil, errors.New("should never reach here")
}

func (fileSystem) Write(file *os.File, value int) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	} else {
		_, err = file.WriteString(writeVal[value&0x1])
		return err
	}
}

func (fileSystem) Read(file *os.File) (res int, err error) {
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	bytesArray := make([]byte, 3)
	var bytesRead int
	bytesRead, err = file.Read(bytesArray)

	if err != nil {
		return 0, err
	}

	value := string(bytesArray[:bytesRead])
	value = strings.TrimSuffix(value, "\n")

	var result int
	if result, err = strconv.Atoi(value); err != nil {
		return 0, err
	} else {
		return result, err
	}
}
