package config

import "testing"

func TestConfig(t *testing.T) {
	c := Config{
		TargetIP:            "192.168.11.2:55555",
		StartupDelaySeconds: 2,
		RunSeconds:          60,
		Tps:                 1,
		Radar: []Radar{
			{
				IsActive: true,
				RadarIP:  "192.168.11.12:55555",
				Types: []MessageType{
					{
						Name:      "ObjectList",
						Directory: "./stressdata/objectlist",
						Run:       false,
					},
				},
			},
		},
	}

	c.SaveToJson("config.json")
}
