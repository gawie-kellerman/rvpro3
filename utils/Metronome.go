package utils

import (
	"math"
	"time"
)

// Metronome class is used to handle the frequency of clicks (timed sequences/instructions)
// Use Start() to initiate the time.
// Use AwaitClick() inside the loop to calculate the time difference, optionally sleep
// and reset the metronome
// Due to rounding, there could be a millisecond increase in metronome
// clicks.
type Metronome struct {
	StartOn       time.Time
	StopOn        time.Time
	Clicks        uint64
	CycleDuration time.Duration

	// IsReal determines whether overflows will result in immediate click (no sleep)
	// or a delayed click.
	// A true metronome guarantees the rhythm even on skipped clicks
	// A false metronome immediately fires the next click if previous beat overran its time
	IsReal bool
}

func (c *Metronome) Start() {
	c.Clicks = 0
	c.StartOn = time.Now()
}

// AwaitClick uses the type configuration to determine the sleep
// period, sleep the processor, reset the metronome and optionally.
// The type also accounts for sleep instruction time deviations
func (c *Metronome) AwaitClick() time.Duration {
	c.StopOn = time.Now()
	c.Clicks++

	duration := c.StopOn.Sub(c.StartOn).Abs()

	if duration < c.CycleDuration {
		sleepFor := time.Duration(math.Abs(float64(duration) - float64(c.CycleDuration)))
		time.Sleep(sleepFor)
		c.StartOn = c.StartOn.Add(c.CycleDuration)
		return sleepFor
	}
	if c.IsReal {
		// 2400 process, 1000 metronome should result in a 60% delay for the
		// next metronome click
		_, fractional := math.Modf(float64(duration) / float64(c.CycleDuration))
		sleepFor := time.Duration((1 - fractional) * float64(c.CycleDuration))
		time.Sleep(sleepFor)

		diff := float64(duration) / float64(c.CycleDuration)
		c.StartOn = c.StopOn.Add(time.Duration(diff + fractional))
		return sleepFor
	}

	c.Start()
	return 0
}
