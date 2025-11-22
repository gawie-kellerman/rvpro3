package main

import (
	"flag"
	"fmt"
	"strings"

	"rvpro3/radarvision.com/utils"
)

type radarUtilCommand int

const (
	cmdHelp radarUtilCommand = iota
	cmdListRadars
	cmdLiveZones
)

type terminal struct {
	indent int
}

func (t *terminal) Indent(delta int) {
	t.indent = t.indent + delta
	t.indent = max(t.indent, 0)
}

func (t *terminal) PrintIndent() {
	fmt.Print(strings.Repeat(" ", t.indent))
}

func (t *terminal) Println(value ...any) {
	t.PrintIndent()
	fmt.Println(value...)
}

func (t *terminal) PrintErrMsg(value ...any) {
	t.PrintIndent()
	fmt.Println(value...)
}

func (t *terminal) PrintErr(err error) {
	t.PrintIndent()
	fmt.Println("Error:", err)
}

func (t *terminal) PrintfLn(format string, a ...any) {
	t.PrintIndent()
	fmt.Printf(format, a...)
	fmt.Println()
}

// PrintfLnKv prints a key value pair followed by a print line
func (t *terminal) PrintfLnKv(key string, format string, a ...any) {
	t.PrintIndent()
	fmt.Printf("%-25s : ", key)
	fmt.Printf(format, a...)
	fmt.Println()
}

func (t *terminal) Print(a ...any) {
	fmt.Print(a...)
}

var Terminal terminal

type radarUtilParams struct {
	clientId               string
	targetIP               string
	command                string
	quitStrategy           string
	quitStrategySeconds    int
	quitStrategyIterations int
	liveConfigFilename     string
}

func (s *radarUtilParams) Setup() {
	clientIdPtr := flag.String("clientid", "0x01000001", "Client ID")
	targetIPPtr := flag.String("targetip", "192.168.11.1:55555", "Target IP")
	commandPtr := flag.String("cmd", "help", "Command to execute.  Options include (help, list-radars, live-zones). E.g. -cmd=live-zones")
	quitStrategyPtr := flag.String("qs", "seconds", "Quit strategy (infinite, iterations, seconds)")
	liveConfigFilenamePtr := flag.String("liveconfig", "live-zones.json", "Path to live config file.")
	flag.IntVar(&s.quitStrategySeconds, "qs-seconds", 10, "seconds=10")
	flag.IntVar(&s.quitStrategyIterations, "qs-iterations", 10, "iterations=10")

	flag.Parse()
	s.clientId = *clientIdPtr
	s.targetIP = *targetIPPtr
	s.command = *commandPtr
	s.liveConfigFilename = *liveConfigFilenamePtr
	s.quitStrategy = *quitStrategyPtr
}

func (s *radarUtilParams) GetClientId() uint32 {
	cid, err := utils.String.ToInt64(s.clientId)
	s.panic(err)

	return uint32(cid)
}

func (s *radarUtilParams) GetTargetIP() utils.IP4 {
	return utils.IP4Builder.FromString(s.targetIP)
}

func (s *radarUtilParams) GetCommand() radarUtilCommand {
	switch s.command {
	case "help":
		return cmdHelp
	case "list-radars":
		return cmdListRadars
	case "live-zones":
		return cmdLiveZones
	default:
		return cmdHelp
	}
}

func (s *radarUtilParams) GetQuitStrategy() QuitStrategy {
	qstType := QuitStrategyType.Parse(QstInfinite, s.quitStrategy)
	res := QuitStrategy{
		Type:          qstType,
		MaxSeconds:    s.quitStrategySeconds,
		MaxIterations: s.quitStrategyIterations,
	}
	return res
}

func (s *radarUtilParams) panic(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *radarUtilParams) GetLiveConfigFilename() string {
	return s.liveConfigFilename
}

func main() {
	fmt.Println("Radar Vision radarutil v 1.0.0 - 20251117")
	params := radarUtilParams{}
	params.Setup()

	switch params.GetCommand() {
	case cmdListRadars:
		cmd := ListRadarsCmd{}
		cmd.Init(&params)
		cmd.Execute()
	case cmdLiveZones:
		cmd := BuildLiveZonesCmd{}
		cmd.Init(&params)
		cmd.Execute()
	default:
		doShowHelp()
	}
}

func doShowHelp() {
	fmt.Println("Usage: radarutil <flags>")
	flag.PrintDefaults()
}
