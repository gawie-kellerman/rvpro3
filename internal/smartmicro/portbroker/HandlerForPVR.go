package portbroker

import "time"

type HandlerForPVR struct {
	RadarChannel *RadarChannel
}

func (h *HandlerForPVR) Init(radarChannel *RadarChannel) {
	h.RadarChannel = radarChannel
}

func (h *HandlerForPVR) Process(now time.Time, payload []byte) {}
