package port

import "rvpro3/radarvision.com/utils"

type PVRReader struct {
	readerMixin
}

func (p *PVRReader) Init(buffer []byte) {
	p.initBuffer(buffer)
}

func (p *PVRReader) IsSupported() bool {
	switch p.VersionMajor {
	case 3:
		return true
	default:
		return false
	}
}

func (p *PVRReader) GetUnixTime() uint32 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU32(p.Buffer, p.Order, p.StartOffset)
	default:
		return 0
	}
}

func (p *PVRReader) GetMilliseconds() uint16 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU16(p.Buffer, p.Order, p.StartOffset+4)
	default:
		return 0
	}
}

func (p *PVRReader) GetNofObjects() uint8 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU8(p.Buffer, p.StartOffset+6)
	default:
		return 0
	}
}

func (p *PVRReader) GetObjectSize() uint8 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU8(p.Buffer, p.StartOffset+7)
	default:
		return 0
	}
}

func (p *PVRReader) GetObjectId(idx int) uint8 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU8(p.Buffer, p.detailOff(idx, 0))
	default:
		return 0
	}
}

func (p *PVRReader) GetObjectClass(idx int) ObjectClassType {
	switch p.VersionMajor {
	case 3:
		return ObjectClassType(utils.OffsetReader.ReadU8(p.Buffer, p.detailOff(idx, 1)))
	default:
		return 0
	}
}

func (p *PVRReader) GetZone(idx int) uint8 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU8(p.Buffer, p.detailOff(idx, 2))
	default:
		return 0
	}
}

func (p *PVRReader) GetCounter(idx int) uint8 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU8(p.Buffer, p.detailOff(idx, 3))
	default:
		return 0
	}
}

func (p *PVRReader) GetSpeed(idx int) float32 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(p.Buffer, p.Order, p.detailOff(idx, 4))
	default:
		return 0
	}
}

func (p *PVRReader) GetHeading(idx int) float32 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(p.Buffer, p.Order, p.detailOff(idx, 8))
	default:
		return 0
	}
}

func (p *PVRReader) GetLength(idx int) float32 {
	switch p.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(p.Buffer, p.Order, p.detailOff(idx, 12))
	default:
		return 0
	}
}

func (p *PVRReader) detailOff(idx int, offset int) int {
	res := p.StartOffset + p.GetHeaderLength() + idx*16 + offset
	return res
}

func (p *PVRReader) GetHeaderLength() int {
	return 8
}

func (p *PVRReader) PrintDetail() {
	utils.Print.Detail("PVR", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("Nof Objects", "%d\n", p.GetNofObjects())
	utils.Print.Detail("Object Size", "%d\n", p.GetObjectSize())
	utils.Print.Detail("Unix Time", "%d\n", p.GetUnixTime())
	utils.Print.Detail("Milliseconds", "%d\n", p.GetMilliseconds())
	utils.Print.Indent(-2)

	for n := 0; n < int(p.GetNofObjects()); n++ {
		utils.Print.Detail("PVR Object ", "%d\n", n)
		utils.Print.Indent(2)
		utils.Print.Detail("Object Id", "%d\n", p.GetObjectId(n))
		utils.Print.Detail("Object Class", "%d, %s\n", p.GetObjectClass(n), p.GetObjectClass(n))
		utils.Print.Detail("Speed", "%f\n", p.GetSpeed(n))
		utils.Print.Detail("Heading", "%f\n", p.GetHeading(n))
		utils.Print.Detail("Length", "%f\n", p.GetLength(n))
		utils.Print.Detail("Counter", "%d\n", p.GetCounter(n))
		utils.Print.Detail("Zone", "%d\n", p.GetZone(n))
		utils.Print.Indent(-2)
	}
}

func (p *PVRReader) TotalSize() int {
	return p.detailOff(int(p.GetNofObjects()), 0)
}
