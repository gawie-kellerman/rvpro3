package port

import (
	"encoding/binary"
	"os"

	"rvpro3/radarvision.com/utils"
)

type PVRHeader struct {
	UnixTime     uint32
	Milliseconds uint16
	ObjectCount  uint8
	ObjectSize   uint8
}

func (h *PVRHeader) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	h.UnixTime = reader.ReadU32(order)
	h.Milliseconds = reader.ReadU16(order)
	h.ObjectCount = reader.ReadU8()
	h.ObjectSize = reader.ReadU8()
}

func (h *PVRHeader) Write(writer *utils.FixedBuffer) {}

type PVRDetail struct {
	ObjectId uint8
	Class    uint8
	Zone     uint8
	Counter  uint8
	Speed    float32
	Heading  float32
	Length   float32
}

func (h *PVRDetail) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	h.ObjectId = reader.ReadU8()
	h.Class = reader.ReadU8()
	h.Zone = reader.ReadU8()
	h.Counter = reader.ReadU8()
	h.Speed = reader.ReadF32(order)
	h.Heading = reader.ReadF32(order)
	h.Length = reader.ReadF32(order)
}

func (h *PVRDetail) Write(writer *utils.FixedBuffer, order binary.ByteOrder) {
	writer.WriteU8(h.ObjectId)
	writer.WriteU8(h.Class)
	writer.WriteU8(h.Zone)
	writer.WriteU8(h.Counter)
	writer.WriteF32(h.Speed, order)
	writer.WriteF32(h.Heading, order)
	writer.WriteF32(h.Length, order)
}

type PVR struct {
	Th       TransportHeader
	Ph       PortHeader
	Header   PVRHeader
	Details  []PVRDetail
	Crc      uint16
	CrcCheck uint16
}

func (h *PVR) ReadPortData(reader *utils.FixedBuffer) {
	order := h.Ph.GetOrder()
	reader.StartReadMarker()

	h.Header.Read(reader, order)
	h.Details = make([]PVRDetail, h.Header.ObjectCount)

	for i := 0; i < int(h.Header.ObjectCount); i++ {
		h.Details[i].Read(reader, order)
	}

	if !h.Th.Flags.IsSkipPayloadCrc() {
		h.CrcCheck = reader.CalcReadCRC()
		h.Crc = reader.ReadU16(order)
	}
}

func (h *PVR) ReadBytes(bytes []byte) error {
	reader := utils.NewFixedBuffer(bytes, 0, len(bytes))
	h.Th.Read(&reader)
	h.Ph.Read(&reader)
	if reader.Err != nil {
		return reader.Err
	}
	h.ReadPortData(&reader)
	return reader.Err
}

func (h *PVR) Write(writer *utils.FixedBuffer) {
	order := h.Ph.GetOrder()
	writer.StartWriteMarker()
	h.Header.ObjectCount = uint8(len(h.Details))
	h.Header.Write(writer)

	for _, detail := range h.Details {
		detail.Write(writer, order)
	}

	if !h.Th.Flags.IsSkipPayloadCrc() {
		h.Crc = writer.WriteCRC16(binary.BigEndian)
		h.CrcCheck = h.Crc
	}
}

func (h *PVR) ReadFile(filename string) (err error) {
	var data []byte

	if data, err = os.ReadFile(filename); err != nil {
		return err
	}

	if err = h.ReadBytes(data); err != nil {
		return err
	}

	return h.Validate()

}

func (h *PVR) Validate() (err error) {
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
