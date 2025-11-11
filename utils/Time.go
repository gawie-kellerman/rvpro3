package utils

import "time"

var tzOffset int
var tzName string

func init() {
	tzName, tzOffset = time.Now().Zone()
	tzOffset *= 1000
}

type timeImpl struct {
}

var Time timeImpl

func (timeImpl) ToLocalMillis(time time.Time) int64 {
	return time.UnixMilli() + int64(tzOffset)
}
