package utils

import (
	"time"
)

var tzOffset int
var tzName string

const (
	DisplayDateTimeMS = "2006-01-02T15:04:05.000"
	DisplayTimeMS     = "15:04:05.000"
	DisplayDate       = "2006-01-02"
	DisplayDateTime   = "2006-01-02T15:04:05"
	DisplayMMDDTime   = "01/02 15:04:05"
	JsonDateTimeMS    = "\"20060102T150405.000\""
)

const (
	FileDate           = "20060102"
	FileDateTimeMS     = "20060102T150405.000"
	FileDateTimeMinute = "20060102T1504"
	FileDateTimeHour   = "20060102T15"
	FileDateTimeSecond = "20060102T150405"
	FileDTNS           = "20060102T150405.000000000"
)

func init() {
	tzName, tzOffset = time.Now().Zone()
	tzOffset *= 1000
}

type timeUtil struct {
}

var Time timeUtil

func (timeUtil) IsExpired(anchor1 time.Time, anchor2 time.Time, duration time.Duration) bool {
	diff := anchor1.Sub(anchor2).Abs()
	return diff > duration
}
