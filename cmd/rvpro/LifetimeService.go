package main

import (
	"sync"
	"time"

	"rvpro3/radarvision.com/utils"
)

const LifetimeServiceName = "Lifetime.Service"

type LifetimeService struct {
	Terminate  bool
	Terminated bool
	Wg         sync.WaitGroup
}

func (l *LifetimeService) InitFromSettings(_ *utils.Settings) {}

func (l *LifetimeService) Start(state *utils.State, _ *utils.Settings) {
	l.Wg.Add(1)
	state.Set(l.GetServiceName(), l)

	go l.run()
}

func (l *LifetimeService) GetServiceName() string {
	return LifetimeServiceName
}

func (l *LifetimeService) StopApplication() {
	l.Wg.Done()
}

func (l *LifetimeService) run() {
	for !l.Terminated {
		time.Sleep(1 * time.Second)
	}
	l.Terminated = true
	l.Wg.Done()
}
