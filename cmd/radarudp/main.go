package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"rvpro3/radarvision.com/internal/smartmicro/service"
)

type RadarServices struct {
	keepAliveService service.UDPKeepAliveService
	listenerService  service.UDPDataServiceOld
	waitGroup        sync.WaitGroup
}

func (s *RadarServices) SetTargetIP(ipAddr string) {
	s.keepAliveService.LocalIPAddr = ipAddr
	s.listenerService.ServerIPAddr = ipAddr
}

func (s *RadarServices) SetClientId(clientId uint32) {
	s.keepAliveService.ClientId = clientId
}

func (s *RadarServices) Execute() {
	s.setup()

	//s.waitGroup.Add(1)
	//go s.keepAliveService.executeReceive()

	s.waitGroup.Add(1)
	go s.listenerService.Execute()

	s.waitGroup.Wait()
}

func (s *RadarServices) setup() {
	s.keepAliveService.Init()
	s.listenerService.Init()

	s.keepAliveService.OnTerminate = func(_ any) {
		s.waitGroup.Done()
	}

	s.listenerService.OnTerminate = func(_ any) {
		s.waitGroup.Done()
	}

	s.keepAliveService.OnError = func(_ any, err error) {
		log.Fatal(err)
	}

	s.listenerService.OnError = func(_ any, err error) {
		log.Fatal(err)
	}

	s.listenerService.OnData = func(us *service.UDPDataServiceOld, u *net.UDPAddr, b []byte) {
		fmt.Printf("%3d bytes received from %v\n", len(b), u)
	}
}

func main() {
	var targetIP string
	var clientId int
	var command string

	flag.StringVar(&targetIP, "targetip", "192.168.11.1:55555", "Local UDP bound target IP address with port")
	flag.IntVar(&clientId, "clientid", 0x100001, "Radar Client ID")
	flag.StringVar(&command, "commands", "radar-list", "List available radars")

	flag.Parse()
	services := RadarServices{}
	services.SetTargetIP(targetIP)
	services.SetClientId(uint32(clientId))
	services.Execute()
}
