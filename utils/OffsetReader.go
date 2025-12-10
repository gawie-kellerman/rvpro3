package utils

import (
	"encoding/binary"
	"math"

	"github.com/howeyc/crc16"
)

// offsetReader assumes that you have done buffer length validation before
// any possible problematic read.  This is to avoid excessive comparison on length
// where the buffer size, and structure size read from it is known
type offsetReader struct {
}

var OffsetReader offsetReader

func (o offsetReader) ReadU8(buffer []byte, offset int) uint8 {
	return buffer[offset]
}

func (o offsetReader) ReadI8(buffer []byte, offset int) int8 {
	return int8(buffer[offset])
}

func (o offsetReader) ReadU16(buffer []byte, order binary.ByteOrder, offset int) uint16 {
	return order.Uint16(buffer[offset : offset+2])
}

func (o offsetReader) ReadU32(buffer []byte, order binary.ByteOrder, offset int) uint32 {
	return order.Uint32(buffer[offset : offset+4])
}

func (o offsetReader) ReadU64(buffer []byte, order binary.ByteOrder, offset int) uint64 {
	return order.Uint64(buffer[offset : offset+8])
}

func (o offsetReader) CopyBytes(buffer []byte, offset int, target []byte) {
	copy(target, buffer[offset:offset+len(target)])
}

func (o offsetReader) ReadI16(buffer []byte, order binary.ByteOrder, offset int) int16 {
	return int16(o.ReadU16(buffer, order, offset))
}

func (o offsetReader) ReadI32(buffer []byte, order binary.ByteOrder, offset int) int32 {
	return int32(o.ReadU32(buffer, order, offset))
}

func (o offsetReader) ReadI64(buffer []byte, order binary.ByteOrder, offset int) int64 {
	return int64(o.ReadU64(buffer, order, offset))
}

func (o offsetReader) ReadF32(buffer []byte, order binary.ByteOrder, offset int) float32 {
	return math.Float32frombits(o.ReadU32(buffer, order, offset))
}

func (o offsetReader) ReadF64(buffer []byte, order binary.ByteOrder, offset int) float64 {
	return math.Float64frombits(o.ReadU64(buffer, order, offset))
}

func (o offsetReader) CalcCRC16(buffer []byte, offset1 int, offset2 int) uint16 {
	return crc16.ChecksumCCITTFalse(buffer[offset1:offset2])
}
