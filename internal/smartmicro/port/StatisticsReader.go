package port

import "rvpro3/radarvision.com/utils"

type StatisticsReader struct {
	readerMixin
}

func (s *StatisticsReader) Init(buffer []byte) {
	s.initBuffer(buffer)
}

func (s *StatisticsReader) IsSupported() bool {
	switch s.VersionMajor {
	case 4:
		return true
	default:
		return false
	}
}

func (s *StatisticsReader) GetNofZones() uint8 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetNofClasses() uint8 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset+1)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetStatusBits() uint8 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset+2)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetActiveFeatures() StatisticsFeatures {
	switch s.VersionMajor {
	case 4:
		return StatisticsFeatures(utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset+3))
	default:
		return 0
	}
}

func (s *StatisticsReader) GetTimestamp() uint32 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU32(s.Buffer, s.Order, s.StartOffset+4)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetMillitime() uint16 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU16(s.Buffer, s.Order, s.StartOffset+8)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetOutputType() StatisticsOutputType {
	switch s.VersionMajor {
	case 4:
		return StatisticsOutputType(utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset+10))
	default:
		return 0
	}
}

func (s *StatisticsReader) GetOutputFormatVersion() uint8 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset+11)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetFrameId() uint16 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU16(s.Buffer, s.Order, s.StartOffset+12)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetFailSafeStatus() StatisticsFailSafe {
	switch s.VersionMajor {
	case 4:
		return StatisticsFailSafe(utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset+14))
	default:
		return 0
	}
}

func (s *StatisticsReader) GetSROVersion() uint8 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(s.Buffer, s.StartOffset+15)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetIntervalCountDown() uint16 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU16(s.Buffer, s.Order, s.StartOffset+16)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetIntervalTime() uint16 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU16(s.Buffer, s.Order, s.StartOffset+18)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetSensorSerial() uint32 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU32(s.Buffer, s.Order, s.StartOffset+20)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetNofStatistics() uint16 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU16(s.Buffer, s.Order, s.StartOffset+24)
	default:
		return 0
	}
}

func (s *StatisticsReader) GetMessageIdx(idx int) uint16 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU16(s.Buffer, s.Order, s.detailOff(idx, 0))
	default:
		return 0
	}
}

func (s *StatisticsReader) GetZone(idx int) uint8 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(s.Buffer, s.detailOff(idx, 2))
	default:
		return 0
	}
}

func (s *StatisticsReader) GetObjectClass(idx int) ObjectClassType {
	switch s.VersionMajor {
	case 4:
		return ObjectClassType(utils.OffsetReader.ReadU8(s.Buffer, s.detailOff(idx, 3)))
	default:
		return 0
	}
}

func (s *StatisticsReader) GetOutput(idx int) uint16 {
	switch s.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU16(s.Buffer, s.Order, s.detailOff(idx, 4))
	default:
		return 0
	}
}

func (s *StatisticsReader) GetMode(idx int) StatisticsMode {
	switch s.VersionMajor {
	case 4:
		return StatisticsMode(utils.OffsetReader.ReadU8(s.Buffer, s.detailOff(idx, 6)))
	default:
		return 0
	}
}

func (s *StatisticsReader) detailOff(idx int, offset int) int {
	// 8 is the detailLen
	res := s.StartOffset + s.GetHeaderLength() + idx*8 + offset
	return res
}

func (s *StatisticsReader) GetHeaderLength() int {
	return 28
}

func (s *StatisticsReader) PrintDetail() {
	utils.Print.Detail("Statistics", "\n")

	utils.Print.Indent(2)
	utils.Print.Detail("Nof Zones", "%d\n", s.GetNofZones())
	utils.Print.Detail("Nof Classes", "%d\n", s.GetNofClasses())
	utils.Print.Detail("Status", "%d\n", s.GetStatusBits())
	utils.Print.Detail("Features", "%d, %s\n", s.GetActiveFeatures(), s.GetActiveFeatures())
	utils.Print.Detail("Timestamp", "%d\n", s.GetTimestamp())
	utils.Print.Detail("Millitime", "%d\n", s.GetMillitime())
	utils.Print.Detail("Output Type", "%d, %s\n", s.GetOutputType(), s.GetOutputType())
	utils.Print.Detail("Output Format Version", "%d\n", s.GetOutputFormatVersion())
	utils.Print.Detail("Frame Id", "%d\n", s.GetFrameId())
	utils.Print.Detail("Fail Safe", "%b, %s\n", s.GetFailSafeStatus(), s.GetFailSafeStatus())
	utils.Print.Detail("SRO Version", "%d\n", s.GetSROVersion())
	utils.Print.Detail("Interval Countdown", "%d\n", s.GetIntervalCountDown())
	utils.Print.Detail("Interval Time", "%d\n", s.GetIntervalTime())
	utils.Print.Detail("Sensor", "%x\n", s.GetSensorSerial())
	utils.Print.Detail("Nof Statistics", "%d\n", s.GetNofStatistics())
	utils.Print.Indent(-2)

	for n := 0; n < int(s.GetNofStatistics()); n++ {
		utils.Print.Detail("Statistic #", "%d\n", n)
		utils.Print.Indent(2)
		utils.Print.Detail("Message Idx", "%d\n", s.GetMessageIdx(n))
		utils.Print.Detail("Zone", "%d\n", s.GetZone(n))
		utils.Print.Detail("Object Class", "%d, %s\n", s.GetObjectClass(n), s.GetObjectClass(n))
		utils.Print.Detail("Output", "%d\n", s.GetOutput(n))
		utils.Print.Detail("Mode", "%d, %s\n", s.GetMode(n), s.GetMode(n))
		utils.Print.Indent(-2)
	}
}

func (s *StatisticsReader) TotalSize() int {
	return s.detailOff(int(s.GetNofStatistics()), 0)
}
