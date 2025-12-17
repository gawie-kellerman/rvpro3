package port

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

const minSize = 12
const startPatternOffset = 0
const protocolVersionOffset = 1
const headerLengthOffset = 2
const payloadLengthOffset = 3
const protocolTypeOffset = 5
const flagsOffset = 6
const flagsDataOffset = 10

type TransportHeaderReader struct {
	Buffer []byte
}

func (t TransportHeaderReader) GetStartPattern() uint8 {
	return utils.OffsetReader.ReadU8(t.Buffer, startPatternOffset)
}

func (t TransportHeaderReader) GetProtocolVersion() uint8 {
	return utils.OffsetReader.ReadU8(t.Buffer, protocolVersionOffset)
}

func (t TransportHeaderReader) GetHeaderLength() uint8 {
	return utils.OffsetReader.ReadU8(t.Buffer, headerLengthOffset)
}

func (t TransportHeaderReader) GetPayloadLength() uint16 {
	return utils.OffsetReader.ReadU16(t.Buffer, binary.BigEndian, payloadLengthOffset)
}

func (t TransportHeaderReader) GetProtocolType() ProtocolType {
	return ProtocolType(utils.OffsetReader.ReadU8(t.Buffer, protocolTypeOffset))
}

func (t TransportHeaderReader) GetFlags() FlagsType {
	return FlagsType(utils.OffsetReader.ReadU32(t.Buffer, binary.BigEndian, flagsOffset))
}

func (t TransportHeaderReader) GetMessageCounter() uint16 {
	flags := t.GetFlags()
	offset := flagsOffset + flags.OffsetOf(FlMessageCount)
	return utils.OffsetReader.ReadU16(t.Buffer, binary.BigEndian, offset)
}

func (t TransportHeaderReader) GetTimeStamp() int64 {
	flags := t.GetFlags()
	offset := flagsOffset + flags.OffsetOf(FlTimestamp)
	return utils.OffsetReader.ReadI64(t.Buffer, binary.BigEndian, offset)
}

func (t TransportHeaderReader) GetSourceClientId() uint32 {
	flags := t.GetFlags()
	offset := flagsDataOffset + flags.OffsetOf(FlSourceClientId)
	return utils.OffsetReader.ReadU32(t.Buffer, binary.BigEndian, offset)
}

func (t TransportHeaderReader) GetTargetClientId() uint32 {
	flags := t.GetFlags()
	offset := flagsDataOffset + flags.OffsetOf(FlTargetClientId)
	return utils.OffsetReader.ReadU32(t.Buffer, binary.BigEndian, offset)
}

func (t TransportHeaderReader) GetDataIdentifier() uint16 {
	flags := t.GetFlags()
	offset := flagsDataOffset + flags.OffsetOf(FlDataIdentifier)
	return utils.OffsetReader.ReadU16(t.Buffer, binary.BigEndian, offset)
}

func (t TransportHeaderReader) GetSegmentation() uint16 {
	flags := t.GetFlags()
	offset := flagsDataOffset + flags.OffsetOf(FlSegmentation)
	return utils.OffsetReader.ReadU16(t.Buffer, binary.BigEndian, offset)
}

func (t TransportHeaderReader) GetCRC16() uint16 {
	return utils.OffsetReader.ReadU16(t.Buffer, binary.BigEndian, int(t.GetHeaderLength()-2))
}

func (t TransportHeaderReader) CalcCRC16() uint16 {
	return utils.OffsetReader.CalcCRC16(t.Buffer, 0, int(t.GetHeaderLength()-2))
}

func (t TransportHeaderReader) CheckCRC() error {
	if t.GetCRC16() != t.CalcCRC16() {
		return ErrHeaderCRC
	}
	return nil
}

func (t TransportHeaderReader) CheckFormat() error {
	if len(t.Buffer) < minSize {
		return ErrTransportHeaderTooSmall
	}

	if len(t.Buffer) < int(t.GetHeaderLength()) {
		return ErrTransportHeaderTooSmall
	}

	if t.GetStartPattern() != StartPattern {
		return ErrTransportHeaderStartPattern
	}

	return nil
}

func (t TransportHeaderReader) PrintDetail() {
	utils.Print.Detail("Transport Header", "\n")
	utils.Print.Indent(2)
	_, _ = utils.Print.Detail("Run Pattern", "0x%x\n", t.GetStartPattern())
	_, _ = utils.Print.Detail("Protocol Version", "%d\n", t.GetProtocolVersion())
	_, _ = utils.Print.Detail("Header Length", "%d\n", t.GetHeaderLength())
	_, _ = utils.Print.Detail("Payload Length", "%d\n", t.GetPayloadLength())
	_, _ = utils.Print.Detail("Protocol Type", "%d, %s\n", t.GetProtocolType(), t.GetProtocolType())
	t.PrintFlags()
	_, _ = utils.Print.Detail("CRC16", "0x%04x\n", t.GetCRC16())
	_, _ = utils.Print.Detail("CRC16 Check", "0x%04x\n", t.CalcCRC16())
	utils.Print.Indent(-2)
}

func (t TransportHeaderReader) PrintFlags() {
	flags := t.GetFlags()

	_, _ = utils.Print.Detail("Flags", "0b%b, %s\n", int(t.GetFlags()), t.GetFlags())

	if flags.IsMessageCount() {
		utils.Print.Detail("Message Data", "%d\n", int(t.GetMessageCounter()))
	}

	if flags.IsTimestamp() {
		utils.Print.Detail("Timestamp", "%d\n", t.GetTimeStamp())
	}

	if flags.IsSourceClientId() {
		utils.Print.Detail("Source Client Id", "%d, %x\n", int(t.GetSourceClientId()), t.GetSourceClientId())
	}

	if flags.IsTargetClientId() {
		utils.Print.Detail("Target Client Id", "%d\n", int(t.GetTargetClientId()))
	}

	if flags.IsDataIdentifier() {
		utils.Print.Detail("data Identifier", "%d\n", int(t.GetDataIdentifier()))
	}

	if flags.IsSegmentation(nil) {
		utils.Print.Detail("Segmentation", "%d\n", int(t.GetSegmentation()))
	}
}

func (t TransportHeaderReader) CombineTo(file1 []byte) []byte {
	res := make([]byte, 0, len(file1)+len(t.Buffer)-int(t.GetHeaderLength()))
	res = append(res, file1...)
	res = append(res, t.Buffer[t.GetHeaderLength():]...)
	return res
}
