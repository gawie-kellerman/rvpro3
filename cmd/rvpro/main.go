package main

import (
	"flag"
	"os"

	"rvpro3/radarvision.com/internal/sdlc/uartsdlc"
	"rvpro3/radarvision.com/internal/smartmicro/broker/udp"
	"rvpro3/radarvision.com/internal/smartmicro/workflows"
	"rvpro3/radarvision.com/utils"
)

var (
	version       string
	buildDate     string
	buildTime     string
	buildCommitID string
)

func showHelp() bool {
	flag.Usage()

	return false
}

func loadArgs() {
	var dir string
	var err error

	gc := &utils.GlobalConfig

	utils.Print.InfoLn("Radar Vision Middleware", version)
	utils.Print.InfoLn("Build Date: ", buildDate, buildTime)
	utils.Print.InfoLn("Build Commit: ", buildCommitID)

	cfgFilename := utils.Args.GetString("--ini|-i", "ini/config.cfg")
	isShowHelp := utils.Args.Has("--help|-h")
	runMode := utils.Args.GetString("--mode|-m", "run")

	if isShowHelp {
		showHelp()
		os.Exit(0)
	}
	if dir, err = os.Getwd(); err != nil {
		utils.Print.ErrorLn("Failed to get current directory", err)
		os.Exit(1)
	}

	if runMode == "defaults" {
		gc.DumpTo(os.Stdout)
		os.Exit(0)
	}

	utils.Print.InfoLn("Using directory", dir)
	utils.Print.InfoLn("Loading config", cfgFilename)

	if err = gc.MergeFromFile(cfgFilename); err != nil {
		utils.Print.ErrorLn(err.Error())
		utils.Print.ErrorLn("Using all default settings")
	}

	overrideArgs()

	if runMode == "merged-defaults" {
		gc.DumpTo(os.Stdout)
		os.Exit(0)
	}
}

func overrideArgs() {
	overrides := utils.Args.GetKVPairIndexes("--override|-o")

	for _, override := range overrides {
		utils.GlobalConfig.SetRaw(
			utils.Args.GetKeyName(override, "--override|-o"),
			utils.Args.GetValue(override),
		)
	}
}

var services []utils.IConfigService

func main() {
	registerServices()
	registerDefaults()
	loadArgs()
	startServices()
	awaitComplete()

	utils.Print.InfoLn("Program completed successfully")
}

func awaitComplete() {
	lts := utils.GlobalState.Get(LifetimeServiceName).(*LifetimeService)
	lts.Wg.Wait()
}

func startServices() {
	utils.Print.InfoLn("Starting services")
	for _, service := range services {
		utils.Print.InfoLn("Starting service", service.GetServiceName())
		service.SetupRunnable(&utils.GlobalState, &utils.GlobalConfig)
	}
}

func registerDefaults() {
	utils.Print.Ln("Registering service defaults")
	for _, service := range services {
		service.SetupDefaults(&utils.GlobalConfig)
	}
}

func registerServices() {
	utils.Print.Ln("Registering services")

	services = make([]utils.IConfigService, 0, 100)
	registerService(new(LifetimeService))
	registerService(new(LoggingService))
	registerService(udp.NewRadarChannels(&workflows.WorkflowBuilder{}))
	registerService(new(uartsdlc.SDLCService))
	registerService(new(uartsdlc.SDLCExecutorService))
	registerService(new(WebService))

	//NB:  When creating RadarChannels, remember to add the WorkflowBuilder
}

func registerService(service utils.IConfigService) {
	services = append(services, service)
}
