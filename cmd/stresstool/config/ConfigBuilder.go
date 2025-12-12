package config

var ConfigBuilder configBuilder

func (configBuilder) CreateSample() Config {
	res := Config{
		TargetIP:            "192.168.11.2:55555",
		StartupDelaySeconds: 2,
		RunSeconds:          10,
		Tps:                 1,
		WebSocketUrl:        "wss://192.168.11.2:443/socket",
		Radar: []Radar{
			{
				IsActive: true,
				RadarIP:  "192.168.11.12:55555",
				Types: []MessageType{
					{
						Name:      "ObjectList",
						Directory: "./_etc/stressdata/objectlist",
						Run:       false,
					},
				},
			},
		},
	}

	return res
}
