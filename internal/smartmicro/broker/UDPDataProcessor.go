package broker

import (
	"errors"
	"net"
	"sync"
	"time"

	errors2 "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

	OnError func(*UDPDataProcessor, error)
}

var ErrInvalidRadar = errors.New("invalid radar")

func (s *UDPDataProcessor) LatchOnto(ds *service.UDPDataServiceOld, waitGroup *sync.WaitGroup) {
	s.hub.Init()
	s.waitGroup = waitGroup
	ds.OnData = s.onDataHandler
}

func (s *UDPDataProcessor) onDataHandler(service *service.UDPDataServiceOld, addr *net.UDPAddr, bytes []byte) {
	radarIndex := RadarIndex(addr)

	if radarIndex == -1 {
		s.logInvalidRadar(addr)
		return
	}

	var th port.TransportHeader
	var ph port.PortHeader

	reader := utils.NewFixedBuffer(bytes, 0, len(bytes))

	th.Read(&reader)
	ph.Read(&reader)

	if reader.Err != nil {
		s.logMappingError(reader.Err)
		return
	}

	if err := th.Validate(); err != nil {
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
	s.onError(err)
}

func (s *UDPDataProcessor) logMappingError(err error) {
	s.onError(err)
}

func (s *UDPDataProcessor) logInvalidRadar(addr *net.UDPAddr) {
	s.onError(errors2.Wrap(ErrInvalidRadar, addr.String()))
}

func (s *UDPDataProcessor) onError(err error) {
	if s.OnError != nil {
		s.OnError(s, err)
	} else {
		log.Err(err)
	}
}
