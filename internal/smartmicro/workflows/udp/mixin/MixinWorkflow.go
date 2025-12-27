package mixin

import (
	"rvpro3/radarvision.com/internal/smartmicro/broker/udp"
)

type MixinWorkflow struct {
	Parent any
}

func (m *MixinWorkflow) GetParent() any {
	return m.Parent
}

func (m *MixinWorkflow) SetParent(parent any) {
	m.Parent = parent
}

func (m *MixinWorkflow) GetChannel() *udp.RadarChannel {
	return m.Parent.(*udp.RadarChannel)
}
