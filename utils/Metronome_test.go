package utils

import (
	"math/rand"
	"testing"
	"time"
)

func TestCycle_SleepMs(t *testing.T) {
	metronome := Metronome{
		CycleDuration: time.Millisecond * 200,
	}

	for i := 0; i < 10; i++ {
		metronome.Start()

		r := rand.Intn(300)
		Print.Ln(i, ". Process is taking", r, "millis")
		time.Sleep(time.Millisecond * time.Duration(r))

		diff := metronome.AwaitClick()
		Print.Ln(i, ". Sleeping for", diff, "millis")
	}
}

func TestCycle_SleepSecs(t *testing.T) {
	metronome := Metronome{
		CycleDuration: time.Second,
		IsReal:        true,
	}

	for i := 0; i < 10; i++ {
		metronome.Start()

		r := rand.Intn(2000)
		Print.Ln(i, ". Process is taking", r, "millis")
		time.Sleep(time.Millisecond * time.Duration(r))

		diff := metronome.AwaitClick()
		Print.Ln(i, ". Sleeping for", diff, "millis")
	}
}
