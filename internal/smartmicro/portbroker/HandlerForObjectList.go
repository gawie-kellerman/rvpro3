package portbroker

import "time"

type HandlerForObjectList struct {
	RadarChannel *RadarChannel
}

func (h *HandlerForObjectList) Init(radarChannel *RadarChannel) {
	h.RadarChannel = radarChannel
}

func (h *HandlerForObjectList) Process(now time.Time, payload []byte) {}
