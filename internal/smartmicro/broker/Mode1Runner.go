package broker

import (
	"log"
	"sync"

	"rvpro3/radarvision.com/internal/smartmicro/service"
)

type Mode1Runner struct {
	KeepAliveService service.UDPKeepAliveService
	ListenerService  service.UDPDataService
	waitGroup        *sync.WaitGroup
	Processor        UDPDataProcessor
}

func (s *Mode1Runner) Execute(wg *sync.WaitGroup) {
	s.setup(s.waitGroup)

	wg.Add(1)
	go s.ListenerService.Execute()

	wg.Add(1)
	go s.KeepAliveService.Execute()
}

func (s *Mode1Runner) setup(group *sync.WaitGroup) {
	s.waitGroup = group

	s.KeepAliveService.Init()
	s.ListenerService.Init()

	s.KeepAliveService.OnTerminate = func(_ any) {
		s.waitGroup.Done()
	}

	s.ListenerService.OnTerminate = func(_ any) {
		s.waitGroup.Done()
	}

	s.KeepAliveService.OnError = func(sender any, err error) {
		log.Fatal(err)
	}

	s.ListenerService.OnError = func(sender any, err error) {
		log.Fatal(err)
	}

	s.Processor.LatchOnto(&s.ListenerService, group)
}
