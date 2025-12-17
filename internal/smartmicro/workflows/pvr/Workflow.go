package pvr

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/portbroker"
)

type Workflow struct {
	radarChannel *portbroker.RadarChannel
}

func (w Workflow) GetRadarChannel() *portbroker.RadarChannel {
	return w.radarChannel
}

func (w Workflow) SetRadarChannel(channel *portbroker.RadarChannel) {
	w.radarChannel = channel
}

func (w Workflow) Process(time time.Time, bytes []byte) {
	//TODO implement me
	panic("implement me")
}
