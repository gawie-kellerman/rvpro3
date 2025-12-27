package main

import (
	"flag"
	"os"

	"rvpro3/radarvision.com/internal/config"
	"rvpro3/radarvision.com/internal/config/globalkey"
	"rvpro3/radarvision.com/utils"
)

func showHelp() bool {
	flag.Usage()

	return false
}

func loadArgs() {
	var dir string
	var err error

	utils.Print.InfoLn("Radar Vision Middleware (rvm) TODO Version")
	utils.Print.InfoLn("Version")

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
		config.RVPro.DumpTo(os.Stdout)
		os.Exit(0)
	}

	utils.Print.InfoLn("Using directory", dir)
	utils.Print.InfoLn("Loading config", cfgFilename)

	if err = config.RVPro.MergeFromFile(cfgFilename); err != nil {
		utils.Print.ErrorLn(err.Error())
		os.Exit(2)
	}

	overrideArgs()

	if runMode == "merged-defaults" {
		config.RVPro.DumpTo(os.Stdout)
		os.Exit(0)
	}
}

func overrideArgs() {
	overrides := utils.Args.GetKVPairIndexes("--override|-o")

	for _, override := range overrides {
		config.RVPro.Set(
			utils.Args.GetKeyName(override, "--override|-o"),
			utils.Args.GetValue(override),
		)
	}
}

func main() {
	loadArgs()

	switch config.RVPro.GlobalStr(globalkey.StartupMode) {
	case globalkey.StartupModeDefault:
		doUDPStartup()
	}

	utils.Print.InfoLn("Program completed successfully")
}
