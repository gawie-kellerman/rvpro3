package main

import (
	"encoding/hex"
	"flag"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"rvpro3/radarvision.com/cmd/sdlcutil/sdlccase"
	"rvpro3/radarvision.com/internal/sdlc"
	"rvpro3/radarvision.com/utils"
)

var portName string
var baudRate int
var dataBits int
var parity serial.Parity
var stopBits serial.StopBits
var runnerName string

const featureAcknowledge = 1

func main() {
	utils.Print.RawLn("Radar Vision")
	utils.Print.RawLn("SDLC Test Utility 1.0.0 - Wed 10 Dec 2025")
	utils.Print.RawLn("RVM Should not run while using this utility!")
	utils.Print.RawLn()

	if !readArgs() {
		return
	}

	service := sdlc.SDLCService{}
	service.Serial.OnConnect = onConnect
	service.Serial.OnError = onSerialError
	service.Serial.OnRead = onSerialRead
	service.Serial.OnWrote = onSerialWrite
	service.OnPopMessage = onPopMessage
	service.Serial.Init(portName, baudRate, dataBits, parity, stopBits)

	wg := sync.WaitGroup{}
	wg.Add(1)
	service.Start()

	time.Sleep(1 * time.Second)

	startRunner(runnerName, &service, func(sdlccase.ISDLCCase) {
		wg.Done()
	})

	wg.Wait()
}

func startRunner(runnerName string, service *sdlc.SDLCService, onTerminate func(sdlccase.ISDLCCase)) {
	switch runnerName {
	case "show-uartfail":
		uart := new(sdlccase.UARTFail)
		runner := uart
		runner.SetService(service)
		runner.Init()
		runner.Start(uart.Execute)

		runner.SetOnTerminate(onTerminate)

	default:
		panic("Unknown runner")
	}
}

func isValidCommand(cmd string) bool {
	switch cmd {
	case "show-uartfail":
		return true

	default:
		return false
	}
}

func readArgs() bool {
	portNameArg := flag.String("port", "/dev/ttymxc2", "serial port device path")
	baudRateArg := flag.Int("baudrate", 115200, "serial baud rate")
	dataBitsArg := flag.Int("databits", 8, "serial data bits")
	stopBitsArg := flag.Int("stopbits", 0, "serial stop bits (default 0)")
	helpArg := flag.Bool("help", false, "show help")
	runnerArg := flag.String("runner", "", "required command to execute")
	maxCyclesArg := flag.Int(sdlccase.MaxCyclesArg, 60, "Maximum number of cycles")
	detectEveryArg := flag.Int(sdlccase.DetectEveryArg, 3, "Send detection every n seconds")
	statusEveryArg := flag.Int(sdlccase.StatusEveryArg, 1, "Status every n seconds")
	cycleDurationArg := flag.Int(sdlccase.CycleDurationArg, 1000, "Cycle duration is n milliseconds")

	flag.Parse()

	portName = *portNameArg
	baudRate = *baudRateArg
	dataBits = *dataBitsArg
	stopBits = serial.StopBits(*stopBitsArg)
	utils.Args.Set(sdlccase.MaxCyclesArg, *maxCyclesArg)
	utils.Args.Set(sdlccase.DetectEveryArg, *detectEveryArg)
	utils.Args.Set(sdlccase.StatusEveryArg, *statusEveryArg)
	utils.Args.Set(sdlccase.CycleDurationArg, *cycleDurationArg)

	runnerName = strings.ToLower(*runnerArg)
	if len(runnerName) == 0 || *helpArg || !isValidCommand(runnerName) {
		flag.Usage()
		utils.Print.RawLn()
		utils.Print.RawLn("Runners:")
		utils.Print.RawLn("  show-uartfail    Shows UART failures by alternating a status and detect request every 3 seconds")
		utils.Print.RawLn("                   Uses: max-cycles, detect-every, status-every, cycle-duration")
		utils.Print.RawLn()
		return false
	}

	return true
}

func onPopMessage(service *sdlc.SDLCService, bytes []byte) {
	utils.Print.Ln("Interpreted Message ", hex.EncodeToString(bytes))

	response := sdlc.SDLCResponseDecoder{}
	utils.Debug.Panic(response.Init(bytes))
	switch response.GetIdentifier() {
	case sdlc.StaticStatusResponseCode:
		obj, err := response.GetStaticStatus()
		utils.Debug.Panic(err)
		obj.PrintDetail()

	case sdlc.DynamicStatusResponseCode:
		obj, err := response.GetDynamicStatus()
		utils.Debug.Panic(err)
		obj.PrintDetail()

	case sdlc.BIUDiagnosticResponseCode:
		obj, err := response.GetBIUDiagnostics()
		utils.Debug.Panic(err)
		obj.PrintDetail()

	case sdlc.SDLCDiagnosticResponseCode:
		obj, err := response.GetSDLCDiagnostics()
		utils.Debug.Panic(err)
		obj.PrintDetail()

	case sdlc.AcknowledgeResponseCode:
		obj, err := response.GetAcknowledge()
		utils.Debug.Panic(err)
		utils.Print.FmtFeature(featureAcknowledge, "Acknowledged %d\n", obj)

	default:
		utils.Print.WarnLn("Unknown SDLC service identifier: ", response.GetIdentifier())
	}
}

func onSerialWrite(connection *sdlc.SerialConnection, bytes []byte) {
	utils.Print.Ln("Serial Wrote: ", hex.EncodeToString(bytes))
}

func onSerialRead(connection *sdlc.SerialConnection, bytes []byte) {
	if len(bytes) == 0 {
		utils.Print.WarnLn("Serial Read: [[No data was read]]")
	} else {
		utils.Print.Ln("Serial Read: ", hex.EncodeToString(bytes))
	}
}

func onSerialError(connection *sdlc.SerialConnection, err error) {
	utils.Print.Ln("Serial error:", err)
}

func onConnect(connection *sdlc.SerialConnection) {
	utils.Print.Ln("Serial Connected")
}
