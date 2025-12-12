package config

type Radar struct {
	IsActive bool          `xml:"isActive,attr"`
	RadarIP  string        `xml:"radarIP,attr"`
	Types    []MessageType `xml:"Type"`
}
