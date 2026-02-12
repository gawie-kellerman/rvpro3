package servicemodel

import (
	"fmt"
	"testing"

	"rvpro3/radarvision.com/utils"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig("config.json")

	utils.Debug.Panic(err)
	fmt.Println(cfg)
}
