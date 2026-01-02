package port

import (
	"rvpro3/radarvision.com/utils"
)

type ObjectListReader struct {
	readerMixin
	DetailLen int
}

func (o *ObjectListReader) Init(buffer []byte) {
	o.initBuffer(buffer)
}

func (o *ObjectListReader) IsSupported() bool {
	var ph PortHeaderReader

	err := o.InitPort(&ph)
	if err != nil {
		return false
	}

	if o.VersionMajor == 3 && o.VersionMinor == 0 && ph.GetHeaderMajorVersion() == 2 && ph.GetPortMinorVersion() == 0 {
		return true
	}

	return false
}

func (o *ObjectListReader) GetHeaderLength() int {
	switch o.VersionMajor {
	case 3:
		return 16
	default:
		return 0
	}
}

func (o *ObjectListReader) GetCycleTime() float32 {
	switch o.VersionMinor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.StartOffset)
	default:
		return 0
	}
}

func (o *ObjectListReader) GetNofObjects() uint16 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU16(o.Buffer, o.Order, o.StartOffset+4)
	default:
		return 0
	}
}

func (o *ObjectListReader) GetRefPoint() ReferencePoint {
	switch o.VersionMinor {
	case 3:
		return ReferencePoint(utils.OffsetReader.ReadU8(o.Buffer, o.StartOffset+6))
	default:
		return 0
	}
}

// GetObjectSize always return 0
func (o *ObjectListReader) GetObjectSize() uint8 {
	switch o.VersionMinor {
	case 3:
		return utils.OffsetReader.ReadU8(o.Buffer, o.StartOffset+7)
	default:
		return 0
	}
}

func (o *ObjectListReader) GetTimestamp() uint64 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU64(o.Buffer, o.Order, o.StartOffset+8)
	default:
		return 0
	}
}

func (o *ObjectListReader) GetPosXFront(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		//return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.StartOffset+16)
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 0))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetPosYFront(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 4))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetPosXFacing(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		//return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.StartOffset+16)
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 8))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetPosYFacing(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		//return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.StartOffset+16)
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 12))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetPosZ(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 16))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetSpeed(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 20))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetHeading(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 24))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetLength(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 28))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetMileage(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 32))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetQuality(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 36))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetAcceleration(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 40))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetObjectId(objIdx int) uint16 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU16(o.Buffer, o.Order, o.detailOff(objIdx, 44))
	default:
		return 0
	}
}

//func (o ObjectListReader) GetIdleCycles(objIdx int) uint16 {
//	switch o.VersionMajor {
//	case 3:
//		return utils.OffsetReader.ReadU16(o.Buffer, o.Order, o.detailOff(objIdx, 38))
//	default:
//		return 0
//	}
//}

func (o *ObjectListReader) GetSplineIndex(objIdx int) uint16 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU16(o.Buffer, o.Order, o.detailOff(objIdx, 40))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetObjectClass(objIdx int) ObjectClassType {
	switch o.VersionMajor {
	case 3:
		return ObjectClassType(utils.OffsetReader.ReadU8(o.Buffer, o.detailOff(objIdx, 46)))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetObjectStatus(objIdx int) uint8 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU8(o.Buffer, o.detailOff(objIdx, 47))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetLane(objIdx int) uint16 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU16(o.Buffer, o.Order, o.detailOff(objIdx, 48))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetCyclesSince(objIdx int) uint16 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU16(o.Buffer, o.Order, o.detailOff(objIdx, 50))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetZone(objIdx int) uint32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadU32(o.Buffer, o.Order, o.detailOff(objIdx, 52))
	default:
		return 0
	}
}

func (o *ObjectListReader) GetHeight(objIdx int) float32 {
	switch o.VersionMajor {
	case 3:
		return utils.OffsetReader.ReadF32(o.Buffer, o.Order, o.detailOff(objIdx, 44))
	default:
		return 0
	}
}

// detailOff cannot depend on object size as it is not working
func (o *ObjectListReader) detailOff(detailNo int, offset int) int {

	res := o.StartOffset + o.GetHeaderLength() + detailNo*o.detailLen() + offset
	return res
}

func (o *ObjectListReader) detailLen() int {
	if o.DetailLen != 0 {
		return o.DetailLen
	}
	o.DetailLen = 96 - 40
	if o.GetRefPoint() == RpFront {
		o.DetailLen += 8
	}

	if o.GetRefPoint() == RpFacingSide || o.GetRefPoint() == RpFacingCorner {
		o.DetailLen += 8
	}
	return o.DetailLen
}

func (o *ObjectListReader) PrintDetail() {
	utils.Print.Detail("Object List Header", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("Metronome Time", "%f\n", o.GetCycleTime())
	utils.Print.Detail("Nof Objects", "%d\n", o.GetNofObjects())
	utils.Print.Detail("Ref Point", "%d, %s\n", o.GetRefPoint(), o.GetRefPoint())
	utils.Print.Detail("Object Size", "%d\n", o.GetObjectSize())
	utils.Print.Detail("Timestamp", "%d\n", o.GetTimestamp())
	utils.Print.Indent(-2)

	for n := 0; n < int(o.GetNofObjects()); n++ {
		utils.Print.Detail("Object #", "%d\n", n+1)
		utils.Print.Indent(2)
		utils.Print.Detail("Object Id", "%d\n", o.GetObjectId(n))
		utils.Print.Detail("Object Class", "%d, %s\n", o.GetObjectClass(n), o.GetObjectClass(n))
		utils.Print.Detail("Lane", "%d\n", o.GetLane(n))
		utils.Print.Detail("Zone", "%d\n", o.GetZone(n))
		utils.Print.Detail("Status", "%d\n", o.GetObjectStatus(n))
		utils.Print.Detail("Pos Front X, Y, Z", "%f,%f\n", o.GetPosXFront(n), o.GetPosYFront(n))
		utils.Print.Detail("Pos Facing X, Y, Z", "%f,%f\n", o.GetPosXFacing(n), o.GetPosYFacing(n))
		utils.Print.Detail("", "Speed: %.1f, Heading: %.1f, Length: %.1f\n", o.GetSpeed(n), o.GetHeading(n), o.GetLength(n))
		utils.Print.Detail("", "Mileage: %.1f, Quality: %.1f, Accel: %.1f\n", o.GetMileage(n), o.GetQuality(n), o.GetAcceleration(n))
		utils.Print.Indent(-2)
	}
}

func (o *ObjectListReader) TotalSize() int {
	return o.detailOff(int(o.GetNofObjects()), 0)
}
