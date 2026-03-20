package state

import (
	"fmt"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
)

type RadarState struct {
	IP               utils.IP4
	SerialStr        string
	Name             string
	IsManualFailSafe bool
	IsAutoFailSafe   bool
	Pipeline         triggerpipeline.TriggerPipeline
	Serial           uint32                               `json:"-"`
	FailSafe         triggerpipeline.ITriggerPipelineItem `json:"-"`
	triggerState     triggerState
}

func (s *RadarState) ReplaceSerial(serial uint32) string {
	if s.Serial != serial {
		s.SerialStr = fmt.Sprintf("%x", serial)
		s.Serial = serial
	}
	return s.SerialStr
}

type radarStateHelper struct {
}

var RadarStateHelper = radarStateHelper{}

func (radarStateHelper) GetStateName(ip utils.IP4) string {
	return "Radar.State-" + ip.String()
}

func (radarStateHelper) Get(ip utils.IP4) *RadarState {
	instance := utils.GlobalState.Get(RadarStateHelper.GetStateName(ip))
	if instance != nil {
		return instance.(*RadarState)
	}
	return nil
}

func (radarStateHelper) GetOrSet(ip utils.IP4) (res *RadarState) {
	stateName := RadarStateHelper.GetStateName(ip)
	if !utils.GlobalState.Has(stateName) {
		res = new(RadarState)
		res = utils.GlobalState.GetOrSet(stateName, res).(*RadarState)
	} else {
		res = utils.GlobalState.Get(stateName).(*RadarState)
	}

	return res

}

type triggerState struct {
	Lo      uint32
	Hi      uint32
	On      time.Time
	Updated bool
}

func (ts *triggerState) Update(on time.Time, lo uint32, hi uint32) bool {
	if ts.Hi != hi || ts.Lo != lo {
		ts.On = on
		ts.Hi = hi
		ts.Lo = lo
		ts.Updated = true
		return true
	}
	ts.Updated = false
	return false
}
