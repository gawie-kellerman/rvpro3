package gpio

import "os"

type File struct {
	Direction Direction
	Port      int
	file      *os.File
}

func (gpf *File) Open() (err error) {
	if gpf.file == nil {
		gpf.file, err = FileSystem.Open(gpf.Port, gpf.Direction)
		return err
	}
	return nil
}

func (gpf *File) Close() {
	if gpf.file != nil {
		_ = gpf.file.Close()
		gpf.file = nil
	}
}

func (gpf *File) Read() (value int, err error) {
	if err = gpf.Open(); err != nil {
		return 0, err
	}
	return FileSystem.Read(gpf.file)
}

func (gpf *File) IsOpen() bool {
	return gpf.file != nil
}

func (gpf *File) Init(direction Direction, port int) {
	gpf.Direction = direction
	gpf.Port = port
}

func (gpf *File) Export() (err error) {
	return FileSystem.ExportPort(gpf.Port)
}

func (gpf *File) AssignDirection(direction Direction) error {
	return FileSystem.WriteDirection(gpf.Port, gpf.Direction)
}
