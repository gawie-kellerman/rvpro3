package port

import (
	"encoding/binary"
	"errors"

	"rvpro3/radarvision.com/utils"
)

const StartPattern = 0x7e
const ProtocolVersion = 1

var ErrCRC = errors.New("CRC error")

type TransportHeader struct {
	StartPattern    uint8
	ProtocolVersion uint8
	//HeaderLength includes the CRC16
	HeaderLength   uint8
	PayloadLength  uint16
	ProtocolType   ProtocolType
	Flags          FlagsType
	MessageCounter uint16
	Timestamp      uint64
	SourceClientId uint32
	TargetClientId uint32
	CRC16          uint16
	CheckCRC16     uint16
}

func (header *TransportHeader) Init() {
	header.StartPattern = StartPattern
	header.ProtocolVersion = 1
	header.HeaderLength = 0x10
	header.ProtocolType = PtSmartMicroPort
}

func (header *TransportHeader) calcHeaderLen() uint8 {
	return 10 + header.Flags.SizeOf(header.Flags)
}

func (header *TransportHeader) Write(writer *utils.FixedBuffer) {
	header.HeaderLength = header.calcHeaderLen()
	writer.StartWriteMarker()
	writer.WriteU8(header.StartPattern)
	writer.WriteU8(header.ProtocolVersion)
	writer.WriteU8(header.HeaderLength)
	writer.WriteU16(header.PayloadLength, binary.BigEndian)
	writer.WriteU8(uint8(header.ProtocolType))
	writer.WriteU32(uint32(header.Flags), binary.BigEndian)

	if header.Flags.IsMessageCount() {
		writer.WriteU16(header.MessageCounter, binary.BigEndian)
	}

	if header.Flags.IsTimestamp() {
		writer.WriteU64(header.Timestamp, binary.BigEndian)
	}

	if header.Flags.IsSourceClientId() {
		writer.WriteU32(header.SourceClientId, binary.BigEndian)
	}

	if header.Flags.IsTargetClientId() {
		writer.WriteU32(header.TargetClientId, binary.BigEndian)
	}
}

func (header *TransportHeader) Read(reader *utils.FixedBuffer) {
	reader.StartReadMarker()
	header.StartPattern = reader.ReadU8()
	header.ProtocolVersion = reader.ReadU8()
	header.HeaderLength = reader.ReadU8()
	header.PayloadLength = reader.ReadU16(binary.BigEndian)
	header.ProtocolType = ProtocolType(reader.ReadU8())
	header.Flags = FlagsType(reader.ReadU32(binary.BigEndian))

	if header.Flags.IsMessageCount() {
		header.MessageCounter = reader.ReadU16(binary.BigEndian)
	}

	if header.Flags.IsTimestamp() {
		header.Timestamp = reader.ReadU64(binary.BigEndian)
	}

	if header.Flags.IsSourceClientId() {
		header.SourceClientId = reader.ReadU32(binary.BigEndian)
	}

	if header.Flags.IsTargetClientId() {
		header.TargetClientId = reader.ReadU32(binary.BigEndian)
	}
	header.CheckCRC16 = reader.CalcReadCRC()
	header.CRC16 = reader.ReadU16(binary.BigEndian)
}

func (header *TransportHeader) IsValid() error {
	if header.CRC16 == header.CheckCRC16 {
		return nil
	}
	return ErrCRC
}
