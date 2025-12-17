package broker

import (
	"log"
	"sync"

	"rvpro3/radarvision.com/internal/smartmicro/service"
)

type Mode1Runner struct {
	KeepAliveService service.UDPKeepAlive
	ListenerService  service.UDPDataServiceOld
	waitGroup        *sync.WaitGroup
	Processor        UDPEventProcessor
}

func (s *Mode1Runner) Execute(wg *sync.WaitGroup) {
	s.setup(s.waitGroup)

	wg.Add(1)
	s.ListenerService.Run()

	wg.Add(1)
	s.KeepAliveService.Run()
}

func (s *Mode1Runner) setup(group *sync.WaitGroup) {
	s.waitGroup = group

	s.KeepAliveService.Init()
	s.ListenerService.Init()

	s.KeepAliveService.OnTerminate = func(_ *service.UDPKeepAlive) {
		s.waitGroup.Done()
	}

	s.ListenerService.OnTerminate = func(_ any) {
		s.waitGroup.Done()
	}

	s.ListenerService.OnError = func(sender any, err error) {
		log.Fatal(err)
	}

	s.Processor.LatchOnto(&s.ListenerService, group)
}
