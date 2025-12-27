package config

import (
	"fmt"
	"os"
	"testing"

	"rvpro3/radarvision.com/internal/config/radarkey"
	"rvpro3/radarvision.com/utils"
)

func TestRVPro_Init(t *testing.T) {
	utils.Debug.Panic(RVPro.SaveToFile("defaults.cfg"))
}

func TestRVPro_Load(t *testing.T) {
	utils.Debug.Panic(RVPro.MergeFromFile("defaults-custom.cfg"))
	RVPro.DumpTo(os.Stdout)
	fmt.Println(RVPro.IndexedStr(Radar, 12, radarkey.ObjectListPath))
	fmt.Println(RVPro.IndexedStr(Radar, 13, radarkey.ObjectListPath))
}
