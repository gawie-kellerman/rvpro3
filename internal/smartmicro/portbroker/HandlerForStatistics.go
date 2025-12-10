package portbroker

import "time"

type HandlerForStatistics struct {
	RadarChannel *RadarChannel
}

func (h *HandlerForStatistics) Init(radarChannel *RadarChannel) {
	h.RadarChannel = radarChannel
}

func (h *HandlerForStatistics) Process(now time.Time, payload []byte) {}
