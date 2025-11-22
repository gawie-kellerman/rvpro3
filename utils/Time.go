package utils

import (
	"time"
)

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

func (timeImpl) IsOlderThan(anchor time.Time, duration time.Duration) bool {
	now := time.Now()
	diff := now.Sub(anchor).Abs()
	return diff > duration
}
