package portbroker

import "time"

type IWorkflow interface {
	GetRadarChannel() *RadarChannel
	SetRadarChannel(*RadarChannel)
	Process(time.Time, []byte)
}
