package config

type MessageType struct {
	Name              string `xml:"name,attr"`
	Directory         string `xml:"directory,attr"`
	Run               bool   `xml:"run,attr"`
	LatchNo           int    `xml:"latchNo,attr"`
	LatchDurationSecs int    `xml:"latchDurationSecs,attr"`
}
