package port

import (
	"strconv"
)

type SectionDetailType byte

const (
	SdtCommand SectionDetailType = iota
	SdtStatus
	SdtParameter
)

type InstructionSectionDetail struct {
	Id       uint16
	Name     string
	DataType InstructionDataType
	Type     SectionDetailType
	Section  *InstructionSection
}

func (s *InstructionSectionDetail) ToDetail() InstructionDetail {
	res := InstructionDetail{
		SectionId:    s.Section.Id,
		DataType:     s.DataType,
		ResponseType: 0,
		RequestType:  ToRequestType(s.Type),
		ParameterId:  s.Id,
		Signature:    s.CalcSign(),
	}

	return res
}

func (s *InstructionSectionDetail) CalcSign() uint32 {
	step1 := s.Section.Name
	step1 += strconv.FormatUint(uint64(s.Section.Id), 10)
	step1 += "0"
	step1Crc := CalcCRC32(step1)

	step2 := s.Name
	step2 += strconv.FormatUint(uint64(s.Id), 10)
	step2 += s.DataType.ToString()
	step2 += strconv.FormatUint(uint64(step1Crc), 10)
	step2 += "0"

	return CalcCRC32(step2)
}

func ToRequestType(dt SectionDetailType) InstructionRequestType {
	switch dt {
	case SdtCommand:
		return ReqTypeCommand

	case SdtParameter:
		return ReqTypeGetParameter

	case SdtStatus:
		return ReqTypeReadStatus
	}
	return 0
}
