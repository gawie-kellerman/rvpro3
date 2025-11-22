package main

import (
	"net"
	"sync"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

type ListRadarsCmd struct {
	clientId     uint32
	targetIP     utils.IP4
	aliveService service.UDPKeepAliveService
	dataService  service.UDPDataService
	waitGroup    sync.WaitGroup
	radarIPs     map[utils.IP4]bool
	quitStrategy QuitStrategy
}

func (s *ListRadarsCmd) Init(params *radarUtilParams) {
	s.quitStrategy = params.GetQuitStrategy()
	s.radarIPs = make(map[utils.IP4]bool)
	s.clientId = params.GetClientId()
	s.targetIP = params.GetTargetIP()

	s.waitGroup.Add(2)
	s.aliveService.Init()

	s.quitStrategy.OnDone = func(strategy *QuitStrategy) {
		Terminal.Println("Quit strategy completed")
		s.aliveService.Stop()
		s.dataService.Stop()
	}

	s.aliveService.OnTerminate = func(sender *service.UDPKeepAliveService) {
		Terminal.Println("Alive service completed")
		s.waitGroup.Done()
	}

	s.dataService.OnTerminate = func(sender *service.UDPDataService) {
		Terminal.Println("Data service completed")
		s.waitGroup.Done()
	}

	s.dataService.OnData = func(dataService *service.UDPDataService, addr net.UDPAddr, bytes []byte) {
		ip4 := utils.IP4Builder.FromIP(addr.IP, addr.Port)
		_, ok := s.radarIPs[ip4]

		if !ok {
			s.radarIPs[ip4] = true
			th := port.TransportHeader{}
			ph := port.PortHeader{}
			reader := utils.NewFixedBuffer(bytes[:], 0, len(bytes))

			th.Read(&reader)
			ph.Read(&reader)

			Terminal.Println("Received data from", addr.String(), "with protocol", th.ProtocolType.ToString())
			s.quitStrategy.Iterate()
		}
	}

	s.dataService.OnError = func(dataService *service.UDPDataService, err error) {
		Terminal.PrintErrMsg("Program abort due to error:")
		Terminal.PrintErr(err)
		s.aliveService.Stop()
		s.dataService.Stop()
	}
}

func (s *ListRadarsCmd) Execute() {
	s.quitStrategy.PrintDetail(&Terminal)
	Terminal.Println("Starting Alive Service")
	Terminal.Indent(2)
	Terminal.PrintfLnKv("Target IP", "%s", s.targetIP.String())
	Terminal.PrintfLnKv("Client ID", "0x%x", s.clientId)
	s.aliveService.Start(s.targetIP)

	Terminal.Indent(-2)
	Terminal.Println("Starting Data Service")
	s.dataService.Start(s.targetIP)

	Terminal.Indent(2)
	Terminal.PrintfLnKv("Listening on", "%s", s.dataService.ListenAddr.String())
	Terminal.Indent(-2)

	s.quitStrategy.Start()

	s.waitGroup.Wait()
	//s.dataService.Stop()
}
