package port

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

type BodyOrder uint8

const LittleEndian = 2
const BigEndian = 1

func (o BodyOrder) ToGo() binary.ByteOrder {
	if o == LittleEndian {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func (o BodyOrder) String() string {
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

func (s *PortHeader) GetByteSize() int {
	return 4 + 2 + 2 + 8 + 4 + 4
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

func (s *PortHeader) Init(identifier PortIdentifier) {
	s.Identifier = identifier
	s.PortMajorVersion = 2
	s.PortMinorVersion = 2
	s.HeaderMajorVersion = 2
	s.HeaderMinorVersion = 0
	s.BodyOrder = LittleEndian
	s.PortIndex = 0
}

func (s *PortHeader) PrintDetail() {
	utils.Print.Detail("Port Header", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("Port Identifier", "%d, 0x%0x, %s\n", s.Identifier, s.Identifier, s.Identifier.String())
	utils.Print.Detail("Port Version", "%d.%d\n", s.PortMajorVersion, s.PortMinorVersion)
	utils.Print.Detail("Timestamp", "%d\n", s.Timestamp)
	utils.Print.Detail("Body Order", "%s\n", s.BodyOrder.String())
	utils.Print.Detail("Port Size", "%d\n", s.PortSize)
	utils.Print.Detail("Header Version", "%d.%d\n", s.HeaderMajorVersion, s.HeaderMinorVersion)
	utils.Print.Indent(-2)
}

func (s *PortHeader) Validate() error {
	return nil
}
