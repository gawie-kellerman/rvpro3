package main

import (
	"flag"
	"os"
	"time"

	"rvpro3/radarvision.com/cmd/stresstool/config"
	"rvpro3/radarvision.com/cmd/stresstool/hive"
	"rvpro3/radarvision.com/utils"
)

var stressStats hive.StressStats

func main() {
	cmdPtr := flag.String("cmd", "", "Command to run.  (run, create-config)")
	configFilePtr := flag.String("config", "config.xml", "XML Configuration file")
	statsFilePtr := flag.String("stats", "rvm-stress.json", "Statistics stress output filename")

	flag.Parse()

	switch *cmdPtr {
	case "run":
		runStress(*configFilePtr, *statsFilePtr)
		return

	case "create-config":
		runCreateConfig(*configFilePtr)
		return

	default:
		showHelp()
	}
}

func fetchRVProCounters(cfg config.Config, msg string) ([]*hive.RVProRadarStat, error) {
	utils.Print.Ln("Fetching KvPairConfigProvider Counters")
	ws := hive.RVProSocket{
		Url: cfg.WebSocketUrl,
	}
	if err := ws.Connect(); err != nil {
		return nil, err
	}
	defer ws.Disconnect()

	if stats, err := ws.ReadRadarStats(); err != nil {
		return nil, err
	} else {
		return stats, nil
	}
}

func runStress(configFilename string, statsFilename string) {
	utils.Print.Ln("Radar Vision")
	utils.Print.Ln("KvPairConfigProvider Stress Tool - Copyright Radar Vision 2025")

	wd, err := os.Getwd()
	utils.Debug.Panic(err)
	utils.Print.Ln("Directory: ", wd)

	utils.Print.Ln("Loading config ", configFilename)
	cfg := config.Config{}
	cfg.LoadFromXml(configFilename)

	simulators := make([]*hive.RadarSimulator, 0, 4)

	counters, err := fetchRVProCounters(cfg, "Fetching Startup KvPairConfigProvider Counters")
	utils.Debug.Panic(err)
	stressStats.CountsBefore.Radar = counters

	stressStats.StartOn = time.Now()

	utils.Print.Ln("Waiting initial detail of", cfg.StartupDelaySeconds, "seconds")
	time.Sleep(time.Duration(cfg.StartupDelaySeconds) * time.Second)

	utils.Print.Ln("Starting Radar Simulators")
	for i := range cfg.Radar {
		radarCfg := &cfg.Radar[i]

		if radarCfg.IsActive {
			simulator := new(hive.RadarSimulator)
			initSimulator(simulator, cfg.ConvertTPSToCoolDownMs(), radarCfg, cfg.TargetIP)
			simulators = append(simulators, simulator)

			utils.Print.Ln("Started Simulator for radar", radarCfg.RadarIP, "to", cfg.TargetIP)
		} else {
			utils.Print.Ln("Skipping Simulator for radar", radarCfg.RadarIP, "to", cfg.TargetIP)
		}
	}

	countdown := cfg.RunSeconds
	for countdown > 0 {
		utils.Print.Ln("Awaiting completion in", countdown, "seconds")
		countdown--
		time.Sleep(time.Duration(1) * time.Second)
	}

	// About to stop all running simulator go routines
	for _, simulator := range simulators {
		simulator.Stop()
		utils.Print.Ln("Stopping simulator for radar", simulator.RadarIP)
	}

	for _, simulator := range simulators {
		simulator.AwaitStop()
		utils.Print.Ln("Stopped simulator for radar", simulator.RadarIP)
	}
	stressStats.EndOn = time.Now()

	// Cooldown
	time.Sleep(time.Duration(1) * time.Second)
	counters, err = fetchRVProCounters(cfg, "Fetching Completion KvPairConfigProvider Counters")
	utils.Debug.Panic(err)
	stressStats.CountsAfter.Radar = counters

	saveStressStats(cfg, simulators, statsFilename)
}

func runCreateConfig(configFilename string) {
	utils.Print.Ln("Creating config...", configFilename)
	cfg := config.ConfigBuilder.CreateSample()
	cfg.SaveToXml(configFilename)
	utils.Print.Ln("Created config")
}

func showHelp() {
	flag.Usage()
}

func saveStressStats(cfg config.Config, simulators []*hive.RadarSimulator, statsFilename string) {
	stressStats.Config = cfg
	stressStats.Simulators = make([]*hive.RadarSimulatorStats, 0, len(simulators))

	for _, simulator := range simulators {
		stressStats.Simulators = append(stressStats.Simulators, simulator.GetStats())
	}

	utils.Print.Ln("Saving stress stats to", statsFilename)
	stressStats.SaveToXml(statsFilename)
}

func initSimulator(simulator *hive.RadarSimulator, cooldownMs int, cfg *config.Radar, targetIPStr string) {
	radarIP := utils.IP4Builder.FromString(cfg.RadarIP)
	targetIP := utils.IP4Builder.FromString(targetIPStr)
	simulator.Init(radarIP, targetIP, cfg.Types, cooldownMs)
	simulator.Start()
}
