package broker

import "time"

type RadarState struct {
	Trigger triggerState
}

type triggerState struct {
	Lo      uint32
	Hi      uint32
	On      time.Time
	Updated bool
}

func (ts *triggerState) Update(on time.Time, lo uint32, hi uint32) bool {
	if ts.Hi != hi || ts.Lo != lo {
		ts.On = on
		ts.Hi = hi
		ts.Lo = lo
		ts.Updated = true
		return true
	}
	ts.Updated = false
	return false
}
