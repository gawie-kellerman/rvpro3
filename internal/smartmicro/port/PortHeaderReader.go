package port

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

const portIdentifierOffset = 0
const portMajorVersionOffset = 4
const portMinorVersionOffset = 6
const timestampOffset = 8
const portSizeOffset = 16
const bodyOrderOffset = 20
const portIndexOffset = 21
const headerMajorOffset = 22
const headerMinorOffset = 23

type PortHeaderReader struct {
	Buffer      []byte
	StartOffset int
}

func (r PortHeaderReader) GetHeaderLength() int {
	return 24
}

func (r PortHeaderReader) GetIdentifier() PortIdentifier {
	return PortIdentifier(utils.OffsetReader.ReadU32(r.Buffer, binary.BigEndian, r.StartOffset+portIdentifierOffset))
}

func (r PortHeaderReader) GetPortMajorVersion() uint16 {
	return utils.OffsetReader.ReadU16(r.Buffer, binary.BigEndian, r.StartOffset+portMajorVersionOffset)
}

func (r PortHeaderReader) GetPortMinorVersion() uint16 {
	return utils.OffsetReader.ReadU16(r.Buffer, binary.BigEndian, r.StartOffset+portMinorVersionOffset)
}

func (r PortHeaderReader) GetTimestamp() int64 {
	return utils.OffsetReader.ReadI64(r.Buffer, binary.BigEndian, r.StartOffset+timestampOffset)
}

func (r PortHeaderReader) GetPortSize() uint32 {
	return utils.OffsetReader.ReadU32(r.Buffer, binary.BigEndian, r.StartOffset+portSizeOffset)
}

func (r PortHeaderReader) GetBodyOrder() BodyOrder {
	return BodyOrder(utils.OffsetReader.ReadU8(r.Buffer, r.StartOffset+bodyOrderOffset))
}

func (r PortHeaderReader) GetPortIndex() uint8 {
	return utils.OffsetReader.ReadU8(r.Buffer, r.StartOffset+portIndexOffset)
}

func (r PortHeaderReader) GetHeaderMajorVersion() uint8 {
	return utils.OffsetReader.ReadU8(r.Buffer, r.StartOffset+headerMajorOffset)
}

func (r PortHeaderReader) GetHeaderMinorVersion() uint8 {
	return utils.OffsetReader.ReadU8(r.Buffer, r.StartOffset+headerMinorOffset)
}

func (r PortHeaderReader) PrintDetail() {
	utils.Print.Detail("Port Header", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("Port Identifier", "%d, 0x%x, %s\n", r.GetIdentifier(), int(r.GetIdentifier()), r.GetIdentifier())
	utils.Print.Detail("Port Version", "%d.%d\n", r.GetPortMajorVersion(), r.GetPortMinorVersion())
	utils.Print.Detail("Timestamp", "%d\n", r.GetTimestamp())
	utils.Print.Detail("Port Size", "%d\n", r.GetPortSize())
	utils.Print.Detail("Body Order", "%d, %s\n", r.GetBodyOrder(), r.GetBodyOrder())
	utils.Print.Detail("Port Index", "%d\n", r.GetPortIndex())
	utils.Print.Detail("Header Version", "%d.%d\n", r.GetHeaderMajorVersion(), r.GetHeaderMinorVersion())
	utils.Print.Indent(-2)
}

func (r PortHeaderReader) Check() error {
	if len(r.Buffer) <= headerMinorOffset+r.StartOffset {
		return ErrPayloadTooSmall
	}
	return nil
}
