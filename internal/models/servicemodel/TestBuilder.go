package servicemodel

type testBuilder struct {
}

var TestBuilder testBuilder

func (testBuilder) Build() *Config {
	res := &Config{
		SiteName:     "Integration Test",
		DistanceUnit: "m",
		SpeedUnit:    "mps",
		Radars: []*Radar{
			{
				RadarIP:         "127.0.0.1:50001",
				RadarName:       "Sim Sensor 1",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "127.0.0.1:50002",
				RadarName:       "Sim Sensor 2",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "127.0.0.1:50003",
				RadarName:       "Sim Sensor 3",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
			{
				RadarIP:         "127.0.0.1:50004",
				RadarName:       "Sim Sensor 4",
				StopBarDistance: "0",
				FailSafeTime:    "30",
				Channels:        []Channel{},
			},
		},
	}

	res.Normalize()
	return res
}
