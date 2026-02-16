package main

import (
	"encoding/json"
	"flag"
	"os"

	"rvpro3/radarvision.com/internal/models/servicemodel"
	"rvpro3/radarvision.com/internal/sdlc/uartsdlc"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/internal/smartmicro/udp/broker"
	"rvpro3/radarvision.com/utils"
)

const startupRunModeSetting = "startup.run.mode"
const startupCfgFileSetting = "startup.cfg.file"

var (
	version       string
	buildDate     string
	buildTime     string
	buildCommitID string
	cfgFilename   string
	runMode       string
)
var services []utils.IRunnableService

func showHelp() bool {
	flag.Usage()

	return false
}

func showBranding() {
	utils.Print.InfoLn("Radar Vision Middleware", version)
	utils.Print.InfoLn("Build Date: ", buildDate, buildTime)
	utils.Print.InfoLn("Build Commit: ", buildCommitID)
}

func showDefaults() {
	utils.GlobalSettings.DumpTo(os.Stdout)
	os.Exit(0)
}

// loadArgs uses the command line arguments to optionally:
// 1. Load the configuration from a file
// 2. Show Help
// 3. Override configuration from the command line
// 4. Run the application or show the defaults
// Important: Unlike rvpro2, rvpro3 does not automatically load a configuration
func loadArgs() *utils.Settings {
	var dir string
	var err error

	cfgFilename = utils.Args.GetString("--ini|-i", "ini/config.cfg")
	isShowHelp := utils.Args.Has("--help|-h")
	runMode = utils.Args.GetString("--mode|-m", "run")

	if isShowHelp {
		showHelp()
		os.Exit(0)
	}

	if dir, err = os.Getwd(); err != nil {
		utils.Print.ErrorLn("Failed to get current directory", err)
		os.Exit(1)
	}

	settings := &utils.Settings{}
	settings.Init()

	settings.SetSettingAsStr(startupRunModeSetting, runMode)
	settings.SetSettingAsStr(startupCfgFileSetting, cfgFilename)

	utils.Print.InfoLn("Using directory ->", dir)
	utils.Print.InfoLn("With config ->", cfgFilename)
	utils.Print.InfoLn("With run mode ->", runMode)

	//if err = settings.MergeFromFile(cfgFilename); err != nil {
	//	utils.Print.ErrorLn(err.Error())
	//	utils.Print.ErrorLn("Using default settings")
	//}

	overrideUsingCmdLine(settings)

	return settings
}

// loadSettingsFile loads the key value pair file
func loadSettingsFile(settings *utils.Settings) *utils.Settings {
	var err error
	var isFile bool

	res := &utils.Settings{}

	fileName := settings.GetOrPutStr(startupCfgFileSetting, "")

	if fileName == "" {
		return res
	}

	if isFile, err = utils.File.Exists(fileName); err != nil {
		goto _errorLabel
	}

	if isFile {
		utils.Print.InfoLn("Loading config from", fileName)

		if err = res.MergeFromFile(fileName); err != nil {
			goto _errorLabel
		}
	} else {
		utils.Print.InfoLn("Config file", fileName, "not found - reverting to default")
	}

	return res

_errorLabel:
	utils.Debug.Panic(err)
	return nil
}

// overrideUsingCmdLine overrides arguments as received from the command line
// The process is:
// 1. Every service stage its onw defauts
// 2. A configuration file override step 1 values
// 3. Command line arguments override step 2 values
// Individual values can be set with the --override or -o flags
// e.g. --override=Sample.Value=abc
func overrideUsingCmdLine(settings *utils.Settings) {
	overrides := utils.Args.GetKVPairIndexes("--override|-o")

	for _, override := range overrides {
		key := utils.Args.GetKeyName(override, "--override|-o")
		_, value := utils.Args.GetPair("--override|-o", key)
		if key != "" {
			settings.SetRaw(key, value)
		}
	}
}

func awaitComplete() {
	lts := utils.GlobalState.Get(LifetimeServiceName).(*LifetimeService)
	lts.Wg.Wait()
}

func startServices() {
	utils.Print.InfoLn("Starting services")
	for _, svc := range services {
		utils.Print.InfoLn("Starting", svc.GetServiceName())
		svc.SetupAndStart(&utils.GlobalState, &utils.GlobalSettings)
	}
}

func registerServiceSettings() *utils.Settings {
	utils.Print.InfoLn("Registering service defaults")

	res := &utils.Settings{}
	res.Init()

	for _, svc := range services {
		svc.SetupDefaults(res)
	}

	return res
}

//func updateServiceSettings(settings *utils.Settings) {
//	utils.Print.Ln("Update service settings")
//
//	res := &utils.Settings{}
//	res.Init()
//
//	for _, svc := range services {
//		svc.InitFromConfig(settings)
//	}
//}

func registerServices(settings *utils.Settings) {
	utils.Print.InfoLn("Registering services")

	services = make([]utils.IRunnableService, 0, 100)
	registerService(new(LifetimeService))
	registerService(new(LoggingService))

	if utils.GlobalSettings.GetOrPutBool("feature.umrr.udp", true) {
		config, err := servicemodel.SettingsBuilder.Build(settings)
		if err != nil {
			utils.Print.ErrorLn("Unable to load channel configuration", err)
			os.Exit(1)
		}
		utils.GlobalState.Set(servicemodel.StateName, config)

		registerService(new(service.UDPKeepAlive))
		registerService(new(service.UDPData))

		// Radar Channels are built using the configuration (json), meaning
		// that any number of radars can be defined.  This also
		// means that the radar port can be different which will be very helpful
		// in integration testing.  The question is however, what is a default config
		registerService(new(broker.UDPBrokersService))
	}

	registerService(new(uartsdlc.SDLCService))
	registerService(new(uartsdlc.SDLCExecutorService))
	registerService(new(WebService))

	//NB:  When creating UDPBrokersService, remember to add the WorkflowBuilder
	//TODO: Add TcpHub/Router back into the fold
	//TODO: SNMP
	//TODO: LCD
}

func registerService(service utils.IRunnableService) {
	services = append(services, service)
}

func main() {
	showBranding()
	args := loadArgs()

	switch runMode {
	case "dump-cmd-settings":
		doDumpCmdSettings(args)
	case "dump-final-settings":
		doDumpFinalSettings(args)

	case "dump-config":
		doDumpTestConfig(args)

	default:
		doRunMode(args)
	}

	// Register all services
}

func doDumpTestConfig(cmdSettings *utils.Settings) {
	fileSettings := loadSettingsFile(cmdSettings)
	cmdSettings.MergeFromSettings(fileSettings)

	if utils.GlobalSettings.GetOrPutBool("feature.umrr.udp", true) {
		config, err := servicemodel.SettingsBuilder.Build(cmdSettings)
		if err != nil {
			utils.Print.ErrorLn("Unable to load channel configuration", err)
			os.Exit(1)
		}

		jsonData, err := json.MarshalIndent(config, "", "  ")
		utils.Debug.Panic(err)
		utils.Print.RawLn(string(jsonData))
	}
	os.Exit(0)
}

func doDumpCmdSettings(args *utils.Settings) {
	args.DumpTo(os.Stdout)
}

func doDumpFinalSettings(args *utils.Settings) {
	fileSettings := loadSettingsFile(args)
	args.MergeFromSettings(fileSettings)

	registerServices(args)
	svcSettings := registerServiceSettings()

	utils.GlobalSettings.MergeFromSettings(svcSettings)
	utils.GlobalSettings.MergeFromSettings(args)
	utils.GlobalSettings.DumpTo(os.Stdout)

	svcSettings = nil
}

func doRunMode(args *utils.Settings) {
	fileSettings := loadSettingsFile(args)
	args.MergeFromSettings(fileSettings)

	registerServices(args)
	svcSettings := registerServiceSettings()

	utils.GlobalSettings.MergeFromSettings(svcSettings)
	utils.GlobalSettings.MergeFromSettings(args)

	svcSettings = nil

	startServices()
	awaitComplete()

	utils.Print.InfoLn("rvm program completed")
}
