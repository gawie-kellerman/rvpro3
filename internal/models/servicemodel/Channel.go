package servicemodel

type Channel struct {
	Channel       int    `json:"Channel"`
	Phase         int    `json:"Phase"`
	MaxHold       string `json:"MaxHold"`
	Extend        string `json:"Extend"`
	FailSafe      string `json:"FailSafe"`
	ChannelSource string `json:"ChannelSource"`
	Zones         []Zone `json:"Zones"`
}
