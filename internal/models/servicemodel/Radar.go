package servicemodel

import (
	"rvpro3/radarvision.com/utils"
)

type Radar struct {
	RadarIP         string    `json:"RadarIP"`
	RadarName       string    `json:"RadarName"`
	StopBarDistance string    `json:"StopBarDistance"`
	FailSafeTime    string    `json:"FailSafeTime"`
	EventLog        string    `json:"EventLog,omitempty"`
	ObjectListLog   string    `json:"ObjectListLog,omitempty"`
	Channels        []Channel `json:"Channels"`

	radarIP utils.IP4
}

func (r *Radar) Normalize() {
	r.radarIP = utils.IP4Builder.FromString(r.RadarIP)
}

func (r *Radar) GetRadarIP() utils.IP4 {
	return r.radarIP
}
