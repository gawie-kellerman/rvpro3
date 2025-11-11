package broker

import (
	"net"
	"sync"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

type UDPDataProcessor struct {
	UsesObjectList  bool
	UsesPVR         bool
	UsesStatistics  bool
	UsesTriggers    bool
	UsesDiagnostics bool
	UsesWgs84       bool
	UsesInstruction bool
	waitGroup       *sync.WaitGroup
	hub             MessageHub
}

func (s *UDPDataProcessor) LatchOnto(ds *service.UDPDataService, waitGroup *sync.WaitGroup) {
	s.hub.Init()
	s.waitGroup = waitGroup
	ds.OnData = s.onDataHandler
}

func (s *UDPDataProcessor) onDataHandler(service *service.UDPDataService, addr *net.UDPAddr, bytes []byte) {
	radarIndex := RadarIndex(addr)

	if radarIndex == -1 {
		s.logInvalidRadar(addr)
		return
	}

	var th port.TransportHeader
	var ph port.PortHeader

	reader := utils.NewFixedBuffer(bytes)

	th.Read(&reader)
	ph.Read(&reader)

	if reader.Err != nil {
		s.logMappingError(reader.Err)
		return
	}

	if err := th.IsValid(); err != nil {
		s.logIntegrityErr(err)
		return
	}

	shouldProcess := s.shouldProcess(ph.Identifier)

	if shouldProcess {
		s.hub.Handle(radarIndex, time.Now(), int(ph.Identifier), bytes)
	}
}

func (s *UDPDataProcessor) shouldProcess(portId port.PortIdentifier) bool {
	switch portId {
	case port.PiObjectList:
		return s.UsesObjectList
	case port.PiPVR:
		return s.UsesPVR
	case port.PiStatistics:
		return s.UsesStatistics
	case port.PiWgs84:
		return s.UsesWgs84
	case port.PiInstruction:
		return s.UsesInstruction
	case port.PiDiagnostics:
		return s.UsesDiagnostics
	case port.PiEventTrigger:
		return s.UsesTriggers
	default:
		return false
	}
}

func (s *UDPDataProcessor) logIntegrityErr(err error) {
}

func (s *UDPDataProcessor) logMappingError(err error) {
}

func (s *UDPDataProcessor) logInvalidRadar(addr *net.UDPAddr) {
}
