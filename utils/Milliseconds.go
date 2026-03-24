package utils

import (
	"encoding/json"
	"time"
)

type Milliseconds time.Duration

func (m Milliseconds) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(m).Milliseconds())
}

func (m Milliseconds) Sleep() {
	time.Sleep(time.Duration(m))
}

func (m Milliseconds) Add(now time.Time) time.Time {
	return now.Add(time.Duration(m))
}

func (m Milliseconds) Expired(now time.Time, previous time.Time) bool {
	millis := now.Sub(previous)
	return millis >= time.Duration(m)
}
