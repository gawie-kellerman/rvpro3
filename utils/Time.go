package utils

import (
	"time"
)

var tzOffset int
var tzName string
var lastTime time.Time

const (
	DisplayDateTimeMS   = "2006-01-02T15:04:05.000"
	DisplayDateTimeZone = "2006-01-02T15:04:05-0700"
	DisplayTimeMS       = "15:04:05.000"
	DisplayDate         = "2006-01-02"
	DisplayDateTime     = "2006-01-02T15:04:05"
	DisplayMMDDTime     = "01/02 15:04:05"
	JsonDateTimeMS      = "\"20060102T150405.000\""
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
	Time.Exact()
}

type timeUtil struct {
}

var Time timeUtil

func (timeUtil) IsExpired(anchor1 time.Time, anchor2 time.Time, duration time.Duration) bool {
	diff := anchor1.Sub(anchor2).Abs()
	return diff > duration
}

func (timeUtil) IsSameDay(date1 time.Time, date2 time.Time) bool {
	var y1, m1, d1 = date1.UTC().Date()
	y2, m2, d2 := date2.UTC().Date()

	return y1 == y2 && m1 == m2 && d1 == d2
}

func (timeUtil) Exact() time.Time {
	lastTime = time.Now()
	return lastTime
}

func (timeUtil) Approx() time.Time {
	return lastTime
}
