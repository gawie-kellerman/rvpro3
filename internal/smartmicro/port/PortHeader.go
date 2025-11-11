package port

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

type PortIdentifier uint32

const PiObjectList = 88
const PiStatistics = 25
const PiDiagnostics = 86
const PiWgs84 = 137
const PiUncertainty = 157
const PiInstruction = 46
const PiEventTrigger = 24
const PiPVR = 29

func (id PortIdentifier) ToString() string {
	switch id {
	case PiObjectList:
		return "ObjectList"
	case PiStatistics:
		return "Statistics"
	case PiDiagnostics:
		return "Diagnostics"
	case PiWgs84:
		return "WGS84"
	case PiUncertainty:
		return "Uncertainty"
	case PiInstruction:
		return "Instruction"
	case PiEventTrigger:
		return "EventTrigger"
	case PiPVR:
		return "PVR"
	default:
		return "Unknown"
	}
}

type BodyOrder uint8

const LittleEndian = 2
const BigEndian = 1

func (o BodyOrder) ToGo() binary.ByteOrder {
	if o == LittleEndian {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func (o BodyOrder) ToString() string {
	switch o {
	case LittleEndian:
		return "Little Endian"
	case BigEndian:
		return "Big Endian"
	default:
		return "Unknown"
	}
}

//goland:noinspection GoNameStartsWithPackageName
type PortHeader struct {
	Identifier         PortIdentifier
	PortMajorVersion   uint16
	PortMinorVersion   uint16
	Timestamp          int64
	PortSize           uint32
	BodyOrder          BodyOrder
	PortIndex          uint8
	HeaderMajorVersion uint8
	HeaderMinorVersion uint8
}

func (s *PortHeader) IsObjectList() bool {
	return s.Identifier == PiObjectList
}

func (s *PortHeader) IsStatistics() bool {
	return s.Identifier == PiStatistics
}

func (s *PortHeader) IsDiagnostics() bool {
	return s.Identifier == PiDiagnostics
}

func (s *PortHeader) IsWgs84() bool {
	return s.Identifier == PiWgs84
}

func (s *PortHeader) IsUncertainty() bool {
	return s.Identifier == PiUncertainty
}

func (s *PortHeader) IsInstruction() bool {
	return s.Identifier == PiInstruction
}

func (s *PortHeader) IsEventTrigger() bool {
	return s.Identifier == PiEventTrigger
}

func (s *PortHeader) IsPVR() bool {
	return s.Identifier == PiPVR
}

func (s *PortHeader) Write(writer *utils.FixedBuffer) {
	writer.WriteU32(uint32(s.Identifier), binary.BigEndian)
	writer.WriteU16(s.PortMajorVersion, binary.BigEndian)
	writer.WriteU16(s.PortMinorVersion, binary.BigEndian)
	writer.WriteU64(uint64(s.Timestamp), binary.BigEndian)
	writer.WriteU32(s.PortSize, binary.BigEndian)
	writer.WriteU8(uint8(s.BodyOrder))
	writer.WriteU8(s.PortIndex)
	writer.WriteU8(s.HeaderMajorVersion)
	writer.WriteU8(s.HeaderMinorVersion)
}

func (s *PortHeader) Read(reader *utils.FixedBuffer) {
	s.Identifier = PortIdentifier(reader.ReadU32(binary.BigEndian))
	s.PortMajorVersion = reader.ReadU16(binary.BigEndian)
	s.PortMinorVersion = reader.ReadU16(binary.BigEndian)
	s.Timestamp = int64(reader.ReadU64(binary.BigEndian))
	s.PortSize = reader.ReadU32(binary.BigEndian)
	s.BodyOrder = BodyOrder(reader.ReadU8())
	s.PortIndex = reader.ReadU8()
	s.HeaderMajorVersion = reader.ReadU8()
	s.HeaderMinorVersion = reader.ReadU8()
}

func (s *PortHeader) GetOrder() binary.ByteOrder {
	return s.BodyOrder.ToGo()
}
