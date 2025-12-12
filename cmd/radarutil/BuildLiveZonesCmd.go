package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

type BuildLiveZonesCmd struct {
	clientId           uint32
	targetIP           utils.IP4
	aliveService       service.UDPKeepAliveService
	dataService        service.UDPDataService
	config             LiveConfig
	insServices        [4]service.InstructionService
	insStarteds        [4]bool
	waitGroup          sync.WaitGroup
	quitStrategy       QuitStrategy
	liveConfig         LiveConfig
	liveConfigFilename string
}

func (s *BuildLiveZonesCmd) Init(params *radarUtilParams) {
	s.quitStrategy = params.GetQuitStrategy()
	s.clientId = params.GetClientId()
	s.targetIP = params.GetTargetIP()
	s.liveConfigFilename = params.GetLiveConfigFilename()
	s.liveConfig.Init()

	s.waitGroup.Add(2)
	s.aliveService.Init()

	for i := range s.insServices {
		insService := &s.insServices[i]
		insService.Init()
		insService.Start(&s.dataService, utils.RadarIPOf(i))
		insService.OnResponse = s.onInstructionResponse
	}

	s.dataService.OnData = s.onDataServiceDataCallback
	s.quitStrategy.OnDone = func(strategy *QuitStrategy) {
		Terminal.Println("Quitting LiveConfig.")
		for i := range s.insServices {
			insService := &s.insServices[i]
			insService.Stop()
		}
		s.aliveService.Stop()
		s.dataService.Stop()
	}

	s.aliveService.OnTerminate = func(aliveService *service.UDPKeepAliveService) {
		s.waitGroup.Add(-1)
	}

	s.dataService.OnTerminate = func(dataService *service.UDPDataService) {
		s.waitGroup.Add(-1)
	}
}

func (s *BuildLiveZonesCmd) onDataServiceDataCallback(ds *service.UDPDataService, addr net.UDPAddr, bytes []byte) {
	radarIP4 := utils.IP4Builder.FromIP(addr.IP, addr.Port)
	radarIndex := utils.RadarIndexOf(radarIP4.ToU32())

	if radarIndex == -1 {
		return
	}

	insService := &s.insServices[radarIndex]

	if !s.insStarteds[radarIndex] {
		s.insStarteds[radarIndex] = true
		ins := port.NewInstruction()
		ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
		ins.Th.SourceClientId = s.clientId
		ins.AddDetail(port.Face08700.Parameters.GetNofZones())
		insService.EnqueueSend(ins, 5)
	}

	th := port.TransportHeader{}
	ph := port.PortHeader{}
	reader := utils.NewFixedBuffer(bytes[:], 0, len(bytes))

	th.Read(&reader)
	ph.Read(&reader)

	if ph.Identifier == port.PiInstruction {
		reader.ResetTo(0, len(bytes))
		ins := &port.Instruction{}

		if err := ins.Read(&reader); err != nil {
			log.Err(err).Msg("Error reading instruction")
			return
		}

		insService.EnqueueReceive(ins)
	}
}

func (s *BuildLiveZonesCmd) onInstructionResponse(insService *service.InstructionService, queueItem *service.SendQueueItem) {
	response := queueItem.Response

	for i := range response.Detail {
		detail := &response.Detail[i]
		radarIP := insService.RadarIP.ToIPString()

		if port.Face08700.Parameters.IsGetNofZones(detail) {
			noZones := int(detail.GetU16(response.Ph.GetOrder()))
			Terminal.Println(noZones, "zones found for", radarIP)
			s.liveConfig.SetupZones(radarIP, noZones)

			for n := 0; n < noZones; n++ {
				ins := port.NewInstruction()
				ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
				ins.Th.SourceClientId = s.clientId
				ins.AddDetail(port.Face08700.Zones.GetNofSegmentsByZone(n))
				insService.EnqueueSend(ins, 5)

				ins = port.NewInstruction()
				ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
				ins.Th.SourceClientId = s.clientId
				ins.AddDetail(port.Face08700.Zones.GetWidthByZone(n))
				insService.EnqueueSend(ins, 6)

				ins = port.NewInstruction()
				ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
				ins.Th.SourceClientId = s.clientId
				ins.AddDetail(port.Face08700.Zones.GetRelayAssignment(n))
				insService.EnqueueSend(ins, 7)
			}
		}

		zoneNo := detail.Element1

		if port.Face08700.Zones.IsNofSegmentsByZone(detail) {
			segments := int(detail.GetU8())
			Terminal.Println(segments, "segments for zone", zoneNo, "radar", radarIP)
			s.liveConfig.SetupSegments(radarIP, int(zoneNo), segments)

			if s.liveConfig.IsSegmentComplete(radarIP) {
				for j := 0; j < s.liveConfig.GetSegmentCount(radarIP); j++ {
					ins := port.NewInstruction()
					ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
					ins.Th.SourceClientId = s.clientId
					ins.AddDetail(port.Face08700.ZoneSegments.GetXSegment(j))
					insService.EnqueueSend(ins, 6)
					ins = port.NewInstruction()
					ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
					ins.Th.SourceClientId = s.clientId
					ins.AddDetail(port.Face08700.ZoneSegments.GetYSegment(j))
					insService.EnqueueSend(ins, 6)
				}
			}
		}

		if port.Face08700.ZoneSegments.IsGetXSegment(detail) {
			segmentNo := detail.Element1
			zoneIx, coordIx := s.liveConfig.ZoneAndCoordBySegment(radarIP, int(segmentNo))
			x := detail.GetF32(response.Ph.GetOrder())
			Terminal.PrintfLn("x %f coordinate for segment %d, mapping to zone %d index %d, radar %s",
				x, segmentNo, zoneIx, coordIx, radarIP)
			s.liveConfig.SetX(radarIP, zoneIx, coordIx, x)
		}

		if port.Face08700.ZoneSegments.IsGetYSegment(detail) {
			segmentNo := detail.Element1
			zoneIx, coordIx := s.liveConfig.ZoneAndCoordBySegment(radarIP, int(segmentNo))
			y := detail.GetF32(response.Ph.GetOrder())
			Terminal.PrintfLn("y %f coordinate for segment %d, mapping to zone %d index %d, radar %s",
				y, segmentNo, zoneIx, coordIx, radarIP)
			s.liveConfig.SetY(radarIP, zoneIx, coordIx, y)
		}

		if port.Face08700.Zones.IsWidthByZone(detail) {
			width := detail.GetF32(response.Ph.GetOrder())
			Terminal.Println(width, "width for zone", zoneNo, "radar", radarIP)
			s.liveConfig.SetWidth(radarIP, int(zoneNo), width)
		}

		if port.Face08700.Zones.IsRelayAssignment(detail) {
			relay := detail.GetU8()
			Terminal.Println("relay", relay, "assigned for zone", zoneNo, "radar", radarIP)
			s.liveConfig.SetTrigger(radarIP, int(zoneNo), int(relay))
		}

		if s.liveConfig.IsComplete() {
			s.quitStrategy.Stop()
		}
	}
}

func (s *BuildLiveZonesCmd) Execute() {
	s.quitStrategy.PrintDetail(&Terminal)

	Terminal.Println("Starting Alive Service")
	Terminal.Indent(2)
	Terminal.PrintfLnKv("Target RVProIP", "%s", s.targetIP.String())
	Terminal.PrintfLnKv("Client ID", "0x%x", s.clientId)
	s.aliveService.Start(s.targetIP)

	Terminal.Indent(-2)
	Terminal.Println("Starting data Service")
	s.dataService.Start(s.targetIP)

	Terminal.Indent(2)
	Terminal.PrintfLnKv("Listening on", "%s", s.dataService.ListenAddr.String())
	Terminal.Indent(-2)

	s.quitStrategy.Start()

	s.waitGroup.Wait()

	if s.liveConfig.IsComplete() {
		Terminal.Println("Live zone complete!")
		Terminal.Println("Saving to", s.liveConfigFilename)
		jsonData, err := json.Marshal(&s.liveConfig)

		if err != nil {
			panic(err)
		}
		if err = os.WriteFile(s.liveConfigFilename, jsonData, 0644); err != nil {
			panic(err)
		}
	} else {
		Terminal.Println("Live zone incomplete!!!")
	}
}

type lifeViewInfo struct {
	RadarIP     utils.IP4
	Zones       []zoneInfo
	assignments int
}

func (z *lifeViewInfo) Init(radarIP utils.IP4) {
	z.RadarIP = radarIP
}

func (z *lifeViewInfo) SetNofZones(i int) {
	z.Zones = make([]zoneInfo, i, i)
}

func (z *lifeViewInfo) GetNoZones() int {
	return cap(z.Zones)
}

func (z *lifeViewInfo) ShouldFetchSegments() bool {
	for i := range z.Zones {
		zone := z.Zones[i]
		if len(zone.segments) == 0 {
			return false
		}
	}
	return true
}

func (z *lifeViewInfo) GetNoSegments() int {
	res := 0
	for i := range z.Zones {
		zone := z.Zones[i]
		res += len(zone.segments)
	}
	return res
}

func (z *lifeViewInfo) SetXSegment(element1 uint16, xValue float32) {
	zone, segment := z.GetZoneSegment(int(element1))
	if zone != nil {
		z.assignments++
		if segment >= len(zone.segments) {
			fmt.Println("got here")
		}
		zone.segments[segment].X = xValue
	}
}

func (z *lifeViewInfo) SetYSegment(element1 uint16, xValue float32) {
	zone, segment := z.GetZoneSegment(int(element1))
	if zone != nil {
		z.assignments++
		if segment >= len(zone.segments) {
			fmt.Println("got here")
		}
		zone.segments[segment].Y = xValue
	}
}

func (z *lifeViewInfo) GetZoneSegment(segmentNo int) (*zoneInfo, int) {
	for i := range z.Zones {
		zone := &z.Zones[i]
		if segmentNo >= len(zone.segments) {
			segmentNo -= len(zone.segments)
		} else {
			return zone, segmentNo
		}
	}

	panic("no zone segment")
}

func (z *lifeViewInfo) GetZoneNoByAssignment(segmentNo int) (int, int) {
	for i := range z.Zones {
		zone := &z.Zones[i]
		if segmentNo >= len(zone.segments) {
			segmentNo -= len(zone.segments)
		} else {
			return i, segmentNo
		}
	}

	panic("no zone segment")
}

func (z *lifeViewInfo) SetWidth(element1 uint16, f32 float32) {
	zone := &z.Zones[element1]
	if zone != nil {
		z.assignments++
		zone.width = f32
	}
}

func (z *lifeViewInfo) SetRelay(element1 uint16, value uint8) {
	zone := &z.Zones[element1]
	if zone != nil {
		z.assignments++
		zone.relay = value
	}
}

func (z *lifeViewInfo) calcAssignments() int {
	if len(z.Zones) == 0 {
		return -1
	}

	res := 0
	for i := range z.Zones {
		zone := &z.Zones[i]
		if len(zone.segments) == 0 {
			return -1
		}

		res += zone.calcAssignments()
	}

	return res
}

func (z *lifeViewInfo) IsComplete() bool {
	return z.calcAssignments() == z.assignments
}

type zoneInfo struct {
	width    float32
	relay    uint8
	segments []zoneSegments
}

func (i *zoneInfo) calcAssignments() int {
	return len(i.segments)*2 + 2
}

func (i *zoneInfo) SetSegments(noSegments int) {
	i.segments = make([]zoneSegments, noSegments)
}

type zoneSegments struct {
	X float32
	Y float32
}
