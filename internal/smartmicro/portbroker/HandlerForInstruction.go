package portbroker

import "time"

type HandlerForInstruction struct {
	RadarChannel *RadarChannel
}

func (h *HandlerForInstruction) Init(radarChannel *RadarChannel)       {}
func (h *HandlerForInstruction) Process(now time.Time, payload []byte) {}
