package portbroker

import "time"

type HandlerForTrigger struct {
	radarChannel *RadarChannel
}

func (h *HandlerForTrigger) Init(radarChannel *RadarChannel) {
	h.radarChannel = radarChannel
}

func (h *HandlerForTrigger) Process(now time.Time, payload []byte) {}
