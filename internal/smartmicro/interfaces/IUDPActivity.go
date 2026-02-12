package interfaces

import "time"

// IUDPActivity is run in the context of a IUDPWorkflow
// It represents a single activity runnable in the context of a
// sensor where the sensor is represented in a data chain of
// IUDPActivity -> IUDPWorkflow -> RadarChannel -> RadarChannels -> UDPData
type IUDPActivity interface {
	GetName() string
	SetName(string)
	Process(IUDPWorkflow, int, time.Time, []byte)
}
