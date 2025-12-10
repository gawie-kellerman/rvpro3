package portbroker

import "time"

type HandlerForDiagnostics struct {
	RadarChannel *RadarChannel
}

func (h *HandlerForDiagnostics) Init(radarChannel *RadarChannel) {
	h.RadarChannel = radarChannel
}

func (h *HandlerForDiagnostics) Process(now time.Time, payload []byte) {
}
