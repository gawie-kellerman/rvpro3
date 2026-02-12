package servicemodel

import (
	"encoding/json"
	"os"
)

type defaultBuilder struct {
}

var DefaultBuilder defaultBuilder

func (defaultBuilder) Build() *Config {
	res := &Config{
		SiteName:     "Site Name",
		DistanceUnit: "ft",
		SpeedUnit:    "mph",
		Radars: []*Radar{
			{
				RadarIP:         "192.168.11.12",
				RadarName:       "Sensor 1",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "192.168.11.13",
				RadarName:       "Sensor 2",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "192.168.11.14",
				RadarName:       "Sensor 3",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "192.168.11.15",
				RadarName:       "Sensor 4",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
		},
	}

	res.Normalize()
	return res
}

func (defaultBuilder) BuildFromFile(fileName string) (*Config, error) {
	data, err := os.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	res := &Config{}
	err = json.Unmarshal(data, res)
	if err == nil {
		res.Normalize()
	}
	return res, err
}
