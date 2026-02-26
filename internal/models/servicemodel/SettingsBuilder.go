package servicemodel

import (
	"strings"

	"rvpro3/radarvision.com/utils"
)

type settingsBuilder struct {
}

var SettingsBuilder settingsBuilder

func (settingsBuilder) Build(gc *utils.Settings) (*Config, error) {
	configFile := gc.GetOrPutStr("startup.cfg.file", "")

	if strings.EqualFold(configFile, "test") {
		return TestBuilder.Build(), nil
	}

	if configFile == "" {
		return DefaultBuilder.Build(), nil
	}

	return DefaultBuilder.BuildFromFile(configFile)
}
