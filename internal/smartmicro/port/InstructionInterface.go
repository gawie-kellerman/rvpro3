package port

type InstructionInterface struct {
	Type     int
	Major    int
	Minor    int
	Patch    int
	Sections map[uint16]*InstructionSection
}

func (i *InstructionInterface) Detail(sectionId uint16, parameterId uint16) InstructionDetail {
	if section, ok := i.Sections[sectionId]; ok {
		if detail, ok := section.Detail[parameterId]; ok {
			return detail.ToDetail()
		}
	}
	return InstructionDetail{}
}

func (i *InstructionInterface) Register(section *InstructionSection) {
	if i.Sections == nil {
		i.Sections = make(map[uint16]*InstructionSection, 10)
	}

	i.Sections[section.Id] = section
}

func (i *InstructionInterface) GetHash() uint32 {
	return hash(byte(i.Type), byte(i.Major), byte(i.Minor), byte(i.Patch))
}

func hash(insType, major, minor, patch byte) (result uint32) {
	result = uint32(insType) << 24
	result |= uint32(major) << 16
	result |= uint32(minor) << 8
	result |= uint32(patch)
	return result
}
