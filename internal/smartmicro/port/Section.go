package port

type InstructionSection struct {
	Id                 uint16
	Name               string
	FormatVersionMajor int
	FormatVersionMinor int
	FormatVersionPatch int
	Detail             map[uint16]*InstructionSectionDetail
}

func (s *InstructionSection) Register(detail *InstructionSectionDetail) {
	if s.Detail == nil {
		s.Detail = make(map[uint16]*InstructionSectionDetail, 10)
	}

	s.Detail[detail.Id] = detail
}
