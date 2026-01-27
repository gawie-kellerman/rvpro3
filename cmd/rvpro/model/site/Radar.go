package site

type Radar struct {
	RadarIP         string    `json:"RadarIP"`
	RadarName       string    `json:"RadarName"`
	StopBarDistance string    `json:"StopBarDistance"`
	FailSafeTime    string    `json:"FailSafeTime"`
	Channels        []Channel `json:"Channels"`
}
