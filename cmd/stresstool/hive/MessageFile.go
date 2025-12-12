package hive

import (
	"os"

	"rvpro3/radarvision.com/utils"
)

type MessageFile struct {
	Filename string `xml:"filename,attr"`
	data     []byte
	Stats    MessageFileStats
}

func (f MessageFile) Bytes() []byte {
	var err error
	if len(f.data) == 0 {
		f.data, err = os.ReadFile(f.Filename)
		utils.Debug.Panic(err)
	}

	return f.data
}
