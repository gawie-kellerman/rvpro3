package eventtrigger

import (
	"math"
	"testing"
)

func TestEventCounters_Count(t *testing.T) {
	ct := EventCounters{}
	ct.Process(0, math.MaxInt32, 0)
	ct.Process(0, math.MaxInt32, 0)
	ct.Process(math.MaxInt32, 0, 0)
	ct.Process(0, math.MaxInt32, 0)
	ct.Process(0, math.MaxInt32, 0)
	ct.Process(math.MaxInt32, 0, 1)
	ct.Dump()
}
