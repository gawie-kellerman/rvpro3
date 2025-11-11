package port

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

type ClientKeepAlive struct {
	StartPattern    uint8
	ProtocolVersion uint8
	HeaderLength    uint8
	PayloadLength   uint16
	ProtocolType    ProtocolType
	Flags           FlagsType
	HeaderCrc       uint16

	MajorVersion uint8
	MinorVersion uint8
	reserved1    uint16
	ClientId     uint32
	TargetIP     uint32
	TargetPort   uint16
	reserved2    uint16
	PayloadCrc   uint16
}

func NewClientKeepAlive(clientId uint32, targetIP uint32, targetPort uint16) ClientKeepAlive {
	return ClientKeepAlive{
		StartPattern:    StartPattern,
		ProtocolVersion: 0x01,
		HeaderLength:    0x0c,
		PayloadLength:   0x10,
		ProtocolType:    PtAliveProtocol,
		ClientId:        clientId,
		TargetIP:        targetIP,
		TargetPort:      targetPort,
		MajorVersion:    3,
		MinorVersion:    0,
	}
}

func (a *ClientKeepAlive) Write(writer *utils.FixedBuffer) {
	writer.StartWriteMarker()
	writer.WriteU8(a.StartPattern)
	writer.WriteU8(a.ProtocolVersion)
	writer.WriteU8(a.HeaderLength)
	writer.WriteU16(a.PayloadLength, binary.BigEndian)
	writer.WriteU8(uint8(a.ProtocolType))
	writer.WriteF32(float32(a.Flags), binary.BigEndian)
	a.HeaderCrc = writer.WriteCRC16(binary.BigEndian)

	writer.StartWriteMarker()
	writer.WriteU8(a.MajorVersion)
	writer.WriteU8(a.MinorVersion)
	writer.WriteU16(a.reserved1, binary.BigEndian)
	writer.WriteU32(a.ClientId, binary.BigEndian)
	writer.WriteU32(a.TargetIP, binary.BigEndian)
	writer.WriteU16(a.TargetPort, binary.BigEndian)
	writer.WriteU16(a.reserved2, binary.BigEndian)
	a.PayloadCrc = writer.WriteCRC16(binary.BigEndian)
}
