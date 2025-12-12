package config

import (
	"encoding/json"
	"encoding/xml"
	"os"

	"rvpro3/radarvision.com/utils"
)

type Config struct {
	TargetIP            string  `xml:"targetIP,attr"`
	StartupDelaySeconds int     `xml:"startupDelaySeconds,attr"`
	RunSeconds          int     `xml:"runSeconds,attr"`
	Tps                 int     `xml:"tps,attr"`
	Radar               []Radar `xml:"Radar"`
	WebSocketUrl        string  `xml:"webSocketUrl,attr"`
}

func (c *Config) SaveToJson(filename string) {
	bytes, err := json.Marshal(c)
	utils.Debug.Panic(err)
	utils.Debug.Panic(os.WriteFile(filename, bytes, 0644))
}

func (c *Config) SaveToXml(filename string) {
	bytes, err := xml.Marshal(c)
	utils.Debug.Panic(err)
	utils.Debug.Panic(os.WriteFile(filename, bytes, 0644))
}

func (c *Config) LoadFromXml(filename string) {
	bytes, err := os.ReadFile(filename)
	utils.Debug.Panic(err)

	err = xml.Unmarshal(bytes, c)
	utils.Debug.Panic(err)
}

func (c *Config) LoadFromJson2(filename string) {
	bytes, err := os.ReadFile(filename)
	utils.Debug.Panic(err)

	err = json.Unmarshal(bytes, c)
	utils.Debug.Panic(err)
}

func (c *Config) ConvertTPSToCoolDownMs() int {
	if c.Tps <= 0 || c.Tps > 1000 {
		c.Tps = 12
	}

	return 1000 / c.Tps
}

type configBuilder struct {
}
