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

func (l *LifetimeService) SetupDefaults(config *utils.Settings) {
}

func (l *LifetimeService) SetupAndStart(state *utils.State, config *utils.Settings) {
	l.Wg.Add(1)
	state.Set(l.GetServiceName(), l)

	go l.Start()
}

func (l *LifetimeService) GetServiceName() string {
	return LifetimeServiceName
}

func (l *LifetimeService) GetServiceNames() []string {
	return nil
}

func (l *LifetimeService) StopApplication() {

}

func (l *LifetimeService) Start() {
	for !l.Terminated {
		time.Sleep(1 * time.Second)
	}
	l.Terminated = true
	l.Wg.Done()
}
