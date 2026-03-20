package trigger

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/internal/smartmicro/udp/state"
)

type StageTriggerActivity struct {
	interfaces.UDPActivityMixin
	PipelineItem triggerpipeline.ITriggerPipelineItem
}

func (s *StageTriggerActivity) Init(workflow interfaces.IUDPWorkflow, index int, fullName string) {
	s.InitBase(workflow, index, fullName)

	radarState := state.RadarStateHelper.GetOrSet(workflow.GetRadarIP())

	if radarState != nil {
		s.PipelineItem = radarState.Pipeline.Find(
			triggerpipeline.Staging,
			workflow.GetRadarIP(),
		)
	}
}

func (s *StageTriggerActivity) Process(time time.Time, bytes []byte) {
	th := port.TransportHeaderReader{
		Buffer: bytes,
	}

	ph := port.PortHeaderReader{
		Buffer:      bytes,
		StartOffset: int(th.GetHeaderLength()),
	}

	if ph.GetPortMajorVersion() == 4 && ph.GetPortMinorVersion() == 0 {
		trigger := port.EventTriggerReader{}
		trigger.Init(bytes)

		// Staging pipeline is mandatory, so PipelineItem can never be nil
		s.PipelineItem.SetTrigger(time, 0, trigger.GetRelays())
	}
}
