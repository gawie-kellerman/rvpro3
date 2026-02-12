package servicemodel

import (
	"encoding/json"
	"os"

	"rvpro3/radarvision.com/utils"
)

const StateName = "umrr.channel.config"

type Config struct {
	SiteName     string   `json:"SiteName"`
	DistanceUnit string   `json:"DistanceUnit"`
	SpeedUnit    string   `json:"SpeedUnit"`
	Radars       []*Radar `json:"Radars"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	config.Normalize()
	return &config, nil
}

func (r *Config) GetRadarByIP(ip4 utils.IP4) *Radar {
	for _, radar := range r.Radars {
		if radar.GetRadarIP().Equals(ip4) {
			return radar
		}
	}
	return nil
}

func (r *Config) GetRadarIndex(ip4 utils.IP4) int {
	for index, radar := range r.Radars {
		radarIP := radar.GetRadarIP()
		if radarIP.Equals(ip4) {
			return index
		}
	}
	return -1
}

func (r *Config) Normalize() {
	for _, radar := range r.Radars {
		radar.Normalize()
	}
}
