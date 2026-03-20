package main

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

type progress int

const (
	GetSimulationMode progress = iota
	GetSimulationModeAwait
	SendSimulationMode
	SendSimulationModeAwait
	Done
	Terminating
)

type RadarStatus struct {
	IP4      utils.IP4
	Progress progress
	Tries    int
	TriedOn  time.Time
}

type RadarStatuses struct {
	List map[utils.IP4]*RadarStatus
}

func (r *RadarStatuses) Init() {
	r.List = make(map[utils.IP4]*RadarStatus)
}

func (rs *RadarStatuses) Get(ip4 utils.IP4) (radar *RadarStatus, isNew bool) {
	radar, ok := rs.List[ip4]

	if !ok {
		radar = new(RadarStatus)
		radar.IP4 = ip4
		radar.Progress = GetSimulationMode
		rs.List[ip4] = radar
	}
	return radar, !ok
}

func (rs *RadarStatuses) IsTerminated() bool {
	for _, radar := range rs.List {
		if radar.Progress != Terminating {
			return false
		}
	}
	return true
}
