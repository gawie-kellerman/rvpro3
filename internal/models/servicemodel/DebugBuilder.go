package servicemodel

type debugBuilder struct {
}

var DebugBuilder debugBuilder

func (debugBuilder) Build() *Config {
	res := &Config{
		SiteName:     "Integration Test",
		DistanceUnit: "m",
		SpeedUnit:    "mps",
		Radars: []*Radar{
			{
				RadarIP:         "192.168.11.12:55555",
				RadarName:       "Sensor 1",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "192.168.11.13:55555",
				RadarName:       "Sensor 2",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "192.168.11.14:55555",
				RadarName:       "Sensor 3",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "192.168.11.15:55555",
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
