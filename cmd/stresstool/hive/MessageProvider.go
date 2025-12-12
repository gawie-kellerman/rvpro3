package hive

import (
	"os"
	"path/filepath"

	"rvpro3/radarvision.com/utils"
)

type MessageProvider struct {
	Directory     string `xml:"directory,attr"`
	IsBlocked     bool   `xml:"-"`
	File          []*MessageFile
	FileIndex     int    `xml:"-"`
	Name          string `xml:"name,attr"`
	LatchNo       int
	LatchDuration int `xml:"-"`
}

func (p *MessageProvider) GetNextFile() *MessageFile {
	if p.IsBlocked || len(p.File) == 0 {
		return nil
	}

	if p.FileIndex >= len(p.File) {
		p.FileIndex = 0
	}

	p.FileIndex++
	return p.File[p.FileIndex-1]
}

func (p *MessageProvider) Init(name string, directory string, latchNo int, latchDurationSecs int) {
	var entries []os.DirEntry
	var err error

	p.Directory = directory
	p.Name = name
	p.File = make([]*MessageFile, 0, 10)
	p.LatchNo = latchNo
	p.LatchDuration = latchDurationSecs

	entries, err = os.ReadDir(directory)
	utils.Debug.Panic(err)

	for _, entry := range entries {
		if !entry.IsDir() {
			fullName := filepath.Join(directory, entry.Name())
			_, err = os.Stat(fullName)
			if err != nil {
				utils.Debug.Panic(err)
			}

			fl := &MessageFile{Filename: fullName}
			fl.Stats.Init()
			p.File = append(p.File, fl)
		}
	}
}

func (p *MessageProvider) IsEOF() bool {
	return p.FileIndex >= len(p.File)
}
