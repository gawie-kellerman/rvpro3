package interfaces

type MixinWorkflow struct {
	Parent any
}

func (m *MixinWorkflow) GetParent() any {
	return m.Parent
}

func (m *MixinWorkflow) SetParent(parent any) {
	m.Parent = parent
}
