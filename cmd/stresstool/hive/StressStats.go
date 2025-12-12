package hive

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"time"

	"rvpro3/radarvision.com/cmd/stresstool/config"
	"rvpro3/radarvision.com/utils"
)

type StressStats struct {
	StartOn      time.Time
	EndOn        time.Time
	Config       config.Config
	Simulators   []*RadarSimulatorStats `xml:"Simulator"`
	CountsBefore radarCounts
	CountsAfter  radarCounts
}

func (s *StressStats) SaveToJson(filename string) {
	body, err := json.Marshal(s)
	utils.Debug.Panic(err)

	err = os.WriteFile(filename, body, 0644)
	utils.Debug.Panic(err)
}

func (s *StressStats) SaveToXml(filename string) {
	body, err := xml.Marshal(s)
	utils.Debug.Panic(err)

	err = os.WriteFile(filename, body, 0644)
	utils.Debug.Panic(err)
}

type radarCounts struct {
	Radar []*RVProRadarStat
}
