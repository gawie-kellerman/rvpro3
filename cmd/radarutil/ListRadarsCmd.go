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
	aliveService service.UDPKeepAlive
	dataService  service.UDPData
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
	s.aliveService.SetupDefaults(&utils.GlobalSettings)
	s.aliveService.InitFromSettings(&utils.GlobalSettings)

	s.quitStrategy.OnDone = func(strategy *QuitStrategy) {
		Terminal.Println("Quit strategy completed")
		s.aliveService.Stop()
		s.dataService.Stop()
	}

	s.aliveService.OnTerminate = func(sender *service.UDPKeepAlive) {
		Terminal.Println("Alive service completed")
		s.waitGroup.Done()
	}

	s.dataService.OnTerminate = func(sender *service.UDPData) {
		Terminal.Println("data service completed")
		s.waitGroup.Done()
	}

	s.dataService.OnData = func(dataService *service.UDPData, addr net.UDPAddr, bytes []byte) {
		ip4 := utils.IP4Builder.FromIP(addr.IP, addr.Port)
		_, ok := s.radarIPs[ip4]

		if !ok {
			s.radarIPs[ip4] = true
			th := port.TransportHeader{}
			ph := port.PortHeader{}
			reader := utils.NewFixedBuffer(bytes[:], 0, len(bytes))

			th.Read(&reader)
			ph.Read(&reader)

			Terminal.Println("Received data from", addr.String(), "with protocol", th.ProtocolType.String())
			s.quitStrategy.Iterate()
		}
	}

	s.dataService.OnError = func(dataService *service.UDPData, err error) {
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
	Terminal.PrintfLnKv("Target RVProIP", "%s", s.targetIP.String())
	Terminal.PrintfLnKv("Client ID", "0x%x", s.clientId)
	s.aliveService.Start()

	Terminal.Indent(-2)
	Terminal.Println("Starting data Service")
	s.dataService.Start()

	Terminal.Indent(2)
	Terminal.PrintfLnKv("Listening on", "%s", s.dataService.ListenAddr.String())
	Terminal.Indent(-2)

	s.quitStrategy.Start()

	s.waitGroup.Wait()
	//s.dataService.Stop()
}
