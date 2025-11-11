package utils

import "time"

type LoopGuard interface {
	ShouldContinue(now time.Time) bool
	GetCycles() uint32
	GetLastOn() time.Time
}

type InfiniteLoopGuard struct {
	Cycles uint32
	LastOn time.Time
}

func (i InfiniteLoopGuard) ShouldContinue(now time.Time) bool {
	i.Cycles++
	i.LastOn = now
	return true
}

func (i InfiniteLoopGuard) GetCycles() uint32 {
	return i.Cycles
}

func (i InfiniteLoopGuard) GetLastOn() time.Time {
	return i.LastOn
}

type TimerLoopGuard struct {
	Cycles    uint32
	LastOn    time.Time
	ExpiresOn time.Time
}

func (s TimerLoopGuard) ShouldContinue(now time.Time) bool {
	s.Cycles++
	s.LastOn = now
	return now.Before(s.ExpiresOn)
}

func (s TimerLoopGuard) GetCycles() uint32 {
	return s.Cycles
}

func (s TimerLoopGuard) GetLastOn() time.Time {
	return s.LastOn
}

type CounterLoopGuard struct {
	Cycles    uint32
	LastOn    time.Time
	MaxCycles uint32
}

func (s CounterLoopGuard) ShouldContinue(now time.Time) bool {
	s.Cycles++
	s.LastOn = now
	return s.MaxCycles < s.Cycles
}

func (s CounterLoopGuard) GetCycles() uint32 {
	return s.Cycles
}

func (s CounterLoopGuard) GetLastOn() time.Time {
	return s.LastOn
}
