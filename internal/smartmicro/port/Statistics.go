package port

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

type StatisticsMode uint8
type StatisticsOutputType uint8
type StatisticsFeatures uint8

const (
	SmVolume StatisticsMode = iota
	SmAvgSpeed
	SmPercSpeed85th
	SmOccupancy
	SmHeadway
	SmGap
	SmUnset
)

const (
	SotCurrentData StatisticsOutputType = iota
	SotArchiveData
)

const (
	SfVolume StatisticsFeatures = 1 << iota
	SfOccupancy
	SfAvgSpeed
	SfPercSpeed85th
	SfHeadway
	SfGap
)

func (s StatisticsMode) ToString() string {
	switch s {
	case SmVolume:
		return "Volume"
	case SmOccupancy:
		return "Occupancy"
	case SmAvgSpeed:
		return "AvgSpeed"
	case SmPercSpeed85th:
		return "PercSpeed85th"
	case SmHeadway:
		return "Headway"
	case SmGap:
		return "Gap"
	case SmUnset:
		return "Unset"
	default:
		return "Unknown"
	}
}

func (s StatisticsOutputType) ToString() string {
	switch s {
	case SotCurrentData:
		return "Current Data"
	case SotArchiveData:
		return "Archive Data"
	default:
		return "Unknown"
	}
}

func (s StatisticsFeatures) ToString() string {
	switch s {
	case SfVolume:
		return "Volume"
	case SfOccupancy:
		return "Occupancy"
	case SfAvgSpeed:
		return "AvgSpeed"
	case SfPercSpeed85th:
		return "PercSpeed85th"
	case SfHeadway:
		return "Headway"
	case SfGap:
		return "Gap"
	default:
		return "Unknown"
	}
}

type Statistics struct {
	Th       TransportHeader
	Ph       PortHeader
	Header   StatisticsHeader
	Details  []StatisticsDetail
	Crc      uint16
	CrcCheck uint16
}

type StatisticsHeader struct {
	NofZones            uint8
	NofClasses          uint8
	StatusBits          uint8
	ActiveFeatures      StatisticsFeatures
	Timestamp           uint32
	Millitime           uint16
	OutputType          StatisticsOutputType
	OutputFormatVersion uint8
	FrameId             uint16
	FailsafeStatus      uint8
	SRO2Version         uint8
	IntervalCountdown   uint16
	IntervalTime        uint16
	SensorSerial        uint32
	NofStatistics       uint16
	StaticPortHeaderPad uint16
}

type StatisticsDetail struct {
	MessageIdx       uint16
	ZoneNo           uint8
	ObjectClass      ObjectClassType
	StatisticsOutput uint16
	Mode             StatisticsMode
	Padding          uint8
}

func (s *StatisticsHeader) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	s.NofZones = reader.ReadU8()
	s.NofClasses = reader.ReadU8()
	s.StatusBits = reader.ReadU8()
	s.ActiveFeatures = StatisticsFeatures(reader.ReadU8())
	s.Timestamp = reader.ReadU32(order)
	s.Millitime = reader.ReadU16(order)
	s.OutputType = StatisticsOutputType(reader.ReadU8())
	s.OutputFormatVersion = reader.ReadU8()
	s.FrameId = reader.ReadU16(order)
	s.FailsafeStatus = reader.ReadU8()
	s.SRO2Version = reader.ReadU8()
	s.IntervalCountdown = reader.ReadU16(order)
	s.IntervalTime = reader.ReadU16(order)
	s.SensorSerial = reader.ReadU32(order)
	s.NofStatistics = reader.ReadU16(order)
	s.StaticPortHeaderPad = reader.ReadU16(order)
}

func (s *StatisticsDetail) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	s.MessageIdx = reader.ReadU16(order)
	s.ZoneNo = reader.ReadU8()
	s.ObjectClass = ObjectClassType(reader.ReadU8())
	s.StatisticsOutput = reader.ReadU16(order)
	s.Mode = StatisticsMode(reader.ReadU8())
	s.Padding = reader.ReadU8()
}

func (s *Statistics) ReadPortData(reader *utils.FixedBuffer) {
	order := s.Ph.GetOrder()
	reader.StartReadMarker()

	s.Header.Read(reader, order)
	s.Details = make([]StatisticsDetail, s.Header.NofStatistics)

	for i := 0; i < int(s.Header.NofStatistics); i++ {
		s.Details[i].Read(reader, order)
	}

	if !s.Th.Flags.IsSkipPayloadCrc() {
		s.CrcCheck = reader.CalcReadCRC()
		s.Crc = reader.ReadU16(binary.BigEndian)
	}
}

func (s *Statistics) ReadBytes(bytes []byte) error {
	reader := utils.NewFixedBuffer(bytes, 0, len(bytes))
	s.Th.Read(&reader)
	s.Ph.Read(&reader)
	if reader.Err != nil {
		return reader.Err
	}
	s.ReadPortData(&reader)
	return reader.Err
}
