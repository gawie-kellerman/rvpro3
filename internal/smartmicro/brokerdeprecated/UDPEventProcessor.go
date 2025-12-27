package brokerdeprecated

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

type UDPEventProcessor struct {
	IsProcessObjectList  bool
	IsProcessPVR         bool
	IsProcessStatistics  bool
	IsProcessTriggers    bool
	IsProcessDiagnostics bool
	IsProcessWgs84       bool
	IsProcessInstruction bool
	waitGroup            *sync.WaitGroup
	hub                  MessageHub

	OnError func(*UDPEventProcessor, error)
}

var ErrInvalidRadar = errors.New("invalid radar")

func (s *UDPEventProcessor) LatchOnto(ds *service.UDPDataServiceOld, waitGroup *sync.WaitGroup) {
	s.hub.Init()
	s.waitGroup = waitGroup
	ds.OnData = s.onDataHandler
}

func (s *UDPEventProcessor) onDataHandler(service *service.UDPDataServiceOld, addr *net.UDPAddr, bytes []byte) {
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

func (s *UDPEventProcessor) shouldProcess(portId port.PortIdentifier) bool {
	switch portId {
	case port.PiObjectList:
		return s.IsProcessObjectList
	case port.PiPVR:
		return s.IsProcessPVR
	case port.PiStatistics:
		return s.IsProcessStatistics
	case port.PiWgs84:
		return s.IsProcessWgs84
	case port.PiInstruction:
		return s.IsProcessInstruction
	case port.PiDiagnostics:
		return s.IsProcessDiagnostics
	case port.PiEventTrigger:
		return s.IsProcessTriggers
	default:
		return false
	}
}

func (s *UDPEventProcessor) logIntegrityErr(err error) {
	s.onError(err)
}

func (s *UDPEventProcessor) logMappingError(err error) {
	s.onError(err)
}

func (s *UDPEventProcessor) logInvalidRadar(addr *net.UDPAddr) {
	s.onError(errors2.Wrap(ErrInvalidRadar, addr.String()))
}

func (s *UDPEventProcessor) onError(err error) {
	if s.OnError != nil {
		s.OnError(s, err)
	} else {
		log.Err(err)
	}
}
