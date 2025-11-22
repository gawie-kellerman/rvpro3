package main

import (
	"strings"
	"time"
)

type QuitStrategyType int

const (
	QstInfinite   QuitStrategyType = iota
	QstSeconds    QuitStrategyType = iota
	QstIterations QuitStrategyType = iota
)

func (qs QuitStrategyType) ToString() string {
	switch qs {
	case QstInfinite:
		return "Infinite"
	case QstSeconds:
		return "Seconds"
	case QstIterations:
		return "Iterations"
	default:
		return "Unknown"
	}
}

func (QuitStrategyType) Parse(source string) QuitStrategyType {
	switch strings.ToLower(source) {
	case "infinite":
		return QstInfinite
	case "seconds":
		return QstSeconds
	case "iterations":
		return QstIterations
	default:
		return QstInfinite
	}
}

type QuitStrategy struct {
	Type QuitStrategyType

	StartOn       time.Time
	Iterations    int
	CycleCooldown time.Duration

	MaxIterations int
	MaxSeconds    int
	OnDone        func(*QuitStrategy)
	terminate     bool
	terminated    bool
}

func (s *QuitStrategy) Start() {
	if s.CycleCooldown == 0 {
		s.CycleCooldown = 1
	}

	s.StartOn = time.Now()
	s.Iterations = 0
	s.terminate = false
	s.terminated = false

	go s.Execute()
}

func (s *QuitStrategy) Execute() {
	for !s.terminate {
		switch s.Type {
		case QstSeconds:
			now := time.Now()
			duration := now.Sub(s.StartOn)

			if int(duration.Seconds()) > s.MaxSeconds {
				s.terminate = true
			}

		case QstIterations:
			if s.Iterations > s.MaxIterations {
				s.terminate = true
			}

		default:

		}

		if !s.terminate {
			time.Sleep(s.CycleCooldown * time.Second)
		}
	}

	if s.OnDone != nil {
		s.OnDone(s)
	}
	s.terminated = true
}

func (s *QuitStrategy) Stop() {
	s.terminate = true

	for !s.terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *QuitStrategy) Iterate() {
	s.Iterations++
}

func (s *QuitStrategy) PrintDetail(t *terminal) {
	t.Print("Quit Strategy is run for ")
	switch s.Type {
	case QstInfinite:
		t.Println("infinite iterations")
	case QstSeconds:
		t.Println(s.MaxSeconds, "seconds")
	case QstIterations:
		t.Println(s.MaxIterations, "iterations")
	default:
		t.Print("Unknown")
	}
}
