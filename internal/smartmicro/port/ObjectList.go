package port

import (
	"encoding/binary"
	"os"

	"rvpro3/radarvision.com/utils"
)

type ObjectListHeader struct {
	CycleDuration    uint32
	NofObjects       uint16
	SelectedRefPoint uint8
	ObjectSize       uint8
	MeasureTimestamp uint64
}

func (h *ObjectListHeader) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	h.CycleDuration = reader.ReadU32(order)
	h.NofObjects = reader.ReadU16(order)
	h.SelectedRefPoint = reader.ReadU8()
	h.ObjectSize = reader.ReadU8()
	h.MeasureTimestamp = reader.ReadU64(order)
}

type ObjectListDetail struct {
	XFront                float32
	YFront                float32
	XFacing               float32
	YFacing               float32
	ZPos                  float32
	Speed                 float32
	Heading               float32
	Length                float32
	Mileage               float32
	Quality               float32
	Acceleration          float32
	Id                    uint16
	Class                 ObjectClassType
	StatusFlags           uint8
	Lane                  uint16
	CyclesSinceLastUpdate uint16
	Zone                  uint32
	WgsLongFront          float64
	WgsLatFront           float64
	WgsLongFacing         float64
	WgsLatFacing          float64
}

func (h *ObjectListDetail) Read(reader *utils.FixedBuffer, objectSize uint8, order binary.ByteOrder) {
	h.XFront = reader.ReadF32(order)
	h.YFront = reader.ReadF32(order)
	h.XFacing = reader.ReadF32(order)
	h.YFacing = reader.ReadF32(order)
	h.ZPos = reader.ReadF32(order)
	h.Speed = reader.ReadF32(order)
	h.Heading = reader.ReadF32(order)
	h.Length = reader.ReadF32(order)
	h.Mileage = reader.ReadF32(order)
	h.Quality = reader.ReadF32(order)
	h.Acceleration = reader.ReadF32(order)
	h.Id = reader.ReadU16(order)
	h.Class = ObjectClassType(reader.ReadU8())
	h.StatusFlags = reader.ReadU8()
	h.Lane = reader.ReadU16(order)
	h.CyclesSinceLastUpdate = reader.ReadU16(order)
	h.Zone = reader.ReadU32(order)

	if objectSize == 96-40 {
		// then we are done
	} else if objectSize == 128-40 {
		h.WgsLongFront = reader.ReadF64(order)
		h.WgsLatFront = reader.ReadF64(order)
		h.WgsLongFacing = reader.ReadF64(order)
		h.WgsLatFacing = reader.ReadF64(order)
	}
}

func (h *ObjectListDetail) IsNew() bool {
	return h.StatusFlags == isNewObject
}

func (h *ObjectListDetail) SetNew(isNew bool) {
	if isNew {
		h.StatusFlags |= 0x01
	} else {
		h.StatusFlags &= ^(uint8(1))
	}
}

type ObjectList struct {
	Th       TransportHeader
	Ph       PortHeader
	Header   ObjectListHeader
	Details  []ObjectListDetail
	CrcCheck uint16
	Crc      uint16
}

func (h *ObjectList) ReadPortData(reader *utils.FixedBuffer) {
	order := h.Ph.GetOrder()
	reader.StartReadMarker()

	h.Header.Read(reader, order)
	h.Details = make([]ObjectListDetail, h.Header.NofObjects)

	for i := 0; i < int(h.Header.NofObjects); i++ {
		h.Details[i].Read(reader, h.Header.ObjectSize, order)
	}

	if !h.Th.Flags.IsSkipPayloadCrc() {
		h.CrcCheck = reader.CalcReadCRC()
		h.Crc = reader.ReadU16(binary.BigEndian)
	}
}

func (h *ObjectList) ReadBytes(bytes []byte) error {
	reader := utils.NewFixedBuffer(bytes, 0, len(bytes))
	h.Th.Read(&reader)
	h.Ph.Read(&reader)
	if reader.Err != nil {
		return reader.Err
	}
	h.ReadPortData(&reader)

	return reader.Err
}

func (h *ObjectList) ReadFile(filename string) (err error) {
	var data []byte

	if data, err = os.ReadFile(filename); err != nil {
		return err
	}

	if err = h.ReadBytes(data); err != nil {
		return err
	}

	return h.Validate()
}

func (h *ObjectList) Validate() (err error) {
	if err = h.Th.Validate(); err != nil {
		return err
	}

	if err = h.Ph.Validate(); err != nil {
		return err
	}

	if h.Crc != h.CrcCheck {
		return ErrPayloadCRC
	}

	return nil
}

func (h *ObjectList) Dump(stdout *os.File) {
}
