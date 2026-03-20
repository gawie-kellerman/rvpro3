package main

import (
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/instruction/face08700"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

type RunMode int

const (
	ShowMode RunMode = iota
	UpdateToDisabled
	UpdateToLines
	UpdateToSplines
)

// ToSmartmicro must correlate to S218SimulationModeEnum
func (rm RunMode) ToSmartmicro() uint32 {
	return uint32(rm) - 1
}

func (rm RunMode) String() string {
	switch rm {
	case ShowMode:
		return "show mode"
	case UpdateToDisabled:
		return "disable (0)"
	case UpdateToLines:
		return "lines (1)"
	case UpdateToSplines:
		return "curves (2)"
	default:
		return "unknown"
	}
}

type StartRadarSim struct {
	aliveService service.UDPKeepAliveService
	dataService  service.UDPDataService
	waitGroup    sync.WaitGroup
	Statuses     RadarStatuses
	RunMode      RunMode
	StartedOn    time.Time
	IsWarnShown  bool
}

func (s *StartRadarSim) InitConfig() {
	gs := &utils.GlobalSettings
	s.aliveService.InitFromSettings(gs)
	s.dataService.InitFromSettings(gs)
	gs.ReadArgs()
}

func (s *StartRadarSim) Init() {
	s.Statuses.Init()

	//s.newSimulationMode = utils.GlobalSettings.Basic.Re
	s.waitGroup.Add(2) // Alive and Data

	s.aliveService.InitFromSettings(&utils.GlobalSettings)
	//s.aliveService.LocalIPAddr = utils.IP4Builder.FromString("192.168.11.102:55555")

	s.dataService.InitFromSettings(&utils.GlobalSettings)
	s.dataService.OnData = s.onDataReceived

	s.aliveService.OnTerminate = func(aliveService *service.UDPKeepAliveService) {
		s.waitGroup.Add(-1)
	}

	s.dataService.OnTerminate = func(dataService *service.UDPDataService) {
		s.waitGroup.Add(-1)
	}
}

func (s *StartRadarSim) Start() {
	utils.Print.Ln("Starting Alive Service")
	s.aliveService.Start(&utils.GlobalState, &utils.GlobalSettings)

	utils.Print.Ln("Starting Data Service")
	s.dataService.Start(&utils.GlobalState, &utils.GlobalSettings)

	s.StartedOn = time.Now()
}

func (s *StartRadarSim) onDataReceived(dataService *service.UDPDataService, addr net.UDPAddr, bytes []byte) {
	ip4 := utils.IP4Builder.FromUDPAddr(addr)

	radar, isNew := s.Statuses.Get(ip4)

	if isNew {
		utils.Print.Ln(radar.IP4, "- Found and Registered New Radar")
	}

	switch radar.Progress {
	case GetSimulationMode:
		utils.Print.Ln(radar.IP4, "- Request Simulation Mode try#", radar.Tries+1)
		ins := s.buildGetSimulationMode()
		payload := ins.SaveAsBytes()
		dataService.WriteData(utils.IP4Builder.FromUDPAddr(addr), payload)
		radar.TriedOn = time.Now()
		radar.Progress += 1

	case GetSimulationModeAwait:
		_, ph := port.Helper.GetHeaders(bytes)

		if ph.GetIdentifier() == port.PiInstruction {
			ins := port.Instruction{}

			if err := ins.ReadBytes(bytes[:]); err != nil {
				utils.Debug.Panic(err)
			}

			det := ins.Find(face08700.TRObjectListSection, face08700.S218SimulationMode)

			if det != nil && det.RequestType == port.ReqTypeGetParameter {
				simMode := face08700.S218SimulationModeEnum(det.GetU32(ins.Ph.GetOrder()))
				utils.Print.Ln(radar.IP4, "- Received Simulation Response:", simMode)

				if s.RunMode == ShowMode {
					radar.Progress = Done
				} else {
					radar.Tries = 0
					radar.Progress += 1
				}
			}
		} else {
			now := time.Now()

			if now.Sub(radar.TriedOn).Seconds() > 5 {
				radar.Tries += 1
				radar.TriedOn = now

				if radar.Tries >= 5 {
					utils.Print.Ln(radar.IP4, " Request Simulation tries (5) exceeded")
					radar.Progress = Terminating
				} else {
					radar.Progress -= 1
				}
			}
		}

	case SendSimulationMode:
		ins := s.buildSetSimulationMode(s.RunMode.ToSmartmicro())
		payload := ins.SaveAsBytes()
		utils.Print.Ln(
			radar.IP4,
			"- Request update simulation mode to",
			s.RunMode.String(),
			" try#",
			radar.Tries+1,
		)
		dataService.WriteData(radar.IP4, payload)
		radar.Progress += 1

	case SendSimulationModeAwait:
		_, ph := port.Helper.GetHeaders(bytes)

		if ph.GetIdentifier() == port.PiInstruction {
			ins := port.Instruction{}

			if err := ins.ReadBytes(bytes[:]); err != nil {
				utils.Debug.Panic(err)
			}

			det := ins.Find(face08700.TRObjectListSection, face08700.S218SimulationMode)

			if det != nil {
				switch det.ResponseType {
				case port.ResTypeSuccess:
					utils.Print.Ln(
						ip4,
						"- Updated simulation mode to",
						face08700.S218SimulationModeEnum(det.GetU32(ins.Ph.GetOrder())),
					)
					radar.Progress += 1

				default:
					utils.Print.ErrorLn(ip4, "- Error Response ", det.ResponseType.ToString())
					radar.Progress = Terminating
				}
			}
		} else {
			now := time.Now()

			if now.Sub(radar.TriedOn).Seconds() > 5 {
				radar.Tries += 1
				radar.TriedOn = now

				if radar.Tries > 5 {
					utils.Print.Ln(radar, "- Update simulation mode time exceeded")
					radar.Progress = Terminating
				} else {
					radar.Progress -= 1
				}
			}
		}

	case Done:
		utils.Print.Ln(ip4, "- Work complete")
		radar.Progress += 1

	case Terminating:
		if s.Statuses.IsTerminated() {
			if time.Now().Sub(s.StartedOn).Seconds() < 3 {
				if !s.IsWarnShown {
					utils.Print.WarnLn("Program too quick... waiting up to 3 seconds for other radars to respond")
					s.IsWarnShown = true
				}
			} else {
				utils.Print.Ln("All Work complete - program terminating successfully")
				s.aliveService.Stop()
				s.dataService.Stop()
			}
		}
	}
}

func (s *StartRadarSim) buildGetSimulationMode() *port.Instruction {
	ins := port.NewInstruction()
	ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
	ins.Th.SourceClientId = s.aliveService.ClientId

	ins.AddDetail(port.InstructionDetail{
		SectionId:    face08700.TRObjectListSection,
		ParameterId:  uint16(face08700.S218SimulationMode),
		DimCount:     0,
		RequestType:  port.ReqTypeGetParameter,
		ResponseType: port.ResTypeNoInstruction,
		DataType:     port.IdtU32,
		Element1:     0,
		Element2:     0,
	})
	return ins
}

func (s *StartRadarSim) buildSetSimulationMode(value uint32) *port.Instruction {
	ins := port.NewInstruction()
	ins.Th.Flags = ins.Th.Flags.Set(port.FlSourceClientId)
	ins.Th.SourceClientId = s.aliveService.ClientId
	ins.AddDetail(port.InstructionDetail{
		SectionId:    face08700.TRObjectListSection,
		ParameterId:  uint16(face08700.S218SimulationMode),
		DimCount:     0,
		RequestType:  port.ReqTypeSetParameter,
		ResponseType: port.ResTypeNoInstruction,
		DataType:     port.IdtU32,
		Element1:     0,
		Element2:     0,
	})
	ins.Detail[0].SetU32(ins.Ph.GetOrder(), value)
	return ins
}

func (s *StartRadarSim) Await() {
	s.waitGroup.Wait()
}

func showHelp() {
	utils.Print.Usage("Usage: ./radar-startsim [command] [options]")
	utils.Print.Usage("Examples:")
	utils.Print.UsageExample("./radar-startsim show-config")
	utils.Print.UsageExample("./radar-startsim show-config -o=udp.keepalive.callbackip=192.168.11.102:55555 -o=udp.keepalive.clientid=0x01000002")
	utils.Print.UsageExample("./radar-startsim start")
	utils.Print.UsageExample("./radar-startsim start -o=udp.keepalive.callbackip=192.168.11.102:55555 -o=udp.keepalive.clientid=0x01000001")
	utils.Print.UsageExample("./radar-startsim stop")
	utils.Print.UsageExample("./radar-startsim stop -o=udp.keepalive.callbackip=192.168.11.102:55555 -o=udp.keepalive.clientid=0x01000001")
	utils.Print.UsageExample("./radar-startsim show-current -o=udp.keepalive.callbackip=192.168.11.102:55555")
	utils.Print.Command("[command]")
	utils.Print.CommandName("help", "Shows this help")

	utils.Print.CommandName("start", "Starts Radar Simulation with lines")
	utils.Print.CommandName("start-splines", "Starts Radar Simulation with splines")
	utils.Print.CommandName("stop", "Stops Radar Simulation")
	utils.Print.CommandName("show-config", "Shows the config used/options")
	utils.Print.CommandName("show-current", "Shows the current radar simulation mode")
}

func main() {
	utils.Print.Title("Radar Vision")
	utils.Print.Title("Start Simulator v1.1.0 - 20260318")

	runner := StartRadarSim{}
	runner.InitConfig()

	mode := strings.ToLower(utils.Args.Command(1, "help"))

	switch mode {
	case "start":
		runner.RunMode = UpdateToLines
		runner.Init()
		runner.Start()
		runner.Await()

	case "start-splines":
		runner.RunMode = UpdateToSplines
		runner.Init()
		runner.Start()
		runner.Await()

	case "stop":
		runner.RunMode = UpdateToDisabled
		runner.Init()
		runner.Start()
		runner.Await()

	case "show-current":
		runner.RunMode = ShowMode
		runner.Init()
		runner.Start()
		runner.Await()

	case "show-config":
		utils.GlobalSettings.DumpTo(os.Stdout)

	default:
		showHelp()
	}
}
