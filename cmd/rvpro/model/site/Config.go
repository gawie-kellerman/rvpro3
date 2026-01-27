package site

import (
	"encoding/json"
	"os"
)

type Config struct {
	SiteName     string  `json:"SiteName"`
	DistanceUnit string  `json:"DistanceUnit"`
	SpeedUnit    string  `json:"SpeedUnit"`
	Radars       []Radar `json:"Radars"`
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
	return &config, nil
}
