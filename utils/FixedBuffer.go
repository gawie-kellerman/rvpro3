package utils

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	"github.com/howeyc/crc16"
)

type FixedBuffer struct {
	buffer      []byte
	WritePos    int
	ReadPos     int
	readMarker  int
	writeMarker int
	Err         error
}

var ErrBufferOverflow = errors.New("buffer overflow")

func NewFixedBuffer(buffer []byte, readPos int, writePos int) FixedBuffer {
	return FixedBuffer{
		buffer:   buffer,
		WritePos: writePos,
		ReadPos:  readPos,
	}
}

func (obj *FixedBuffer) AvailForWrite() int {
	return cap(obj.buffer) - obj.WritePos
}

func (obj *FixedBuffer) CanWrite(needBytes int) bool {
	newOff := obj.WritePos + needBytes

	return newOff <= cap(obj.buffer)
}

func (obj *FixedBuffer) CanRead(nofBytes int) bool {
	newOff := obj.ReadPos + nofBytes

	return newOff <= obj.WritePos
}

func (obj *FixedBuffer) WriteBytes(write []byte) {
	if !obj.CanWrite(len(write)) {
		obj.Err = ErrBufferOverflow
		return
	}

	obj.WritePos += copy(obj.buffer[obj.WritePos:obj.WritePos+len(write)], write)
}

func (obj *FixedBuffer) Reset() {
	obj.Err = nil
	obj.buffer = obj.buffer[:0]
	obj.ReadPos = 0
	obj.WritePos = 0
	obj.readMarker = 0
	obj.writeMarker = 0
}

func (obj *FixedBuffer) WriteU8(value uint8) {
	if !obj.CanWrite(1) {
		obj.Err = ErrBufferOverflow
		return
	}

	obj.buffer[obj.WritePos] = value
	obj.WritePos++
}

func (obj *FixedBuffer) ReadU8() uint8 {
	if !obj.CanRead(1) {
		obj.Err = ErrBufferOverflow
		return 0
	}

	result := obj.buffer[obj.ReadPos]
	obj.ReadPos++
	return result
}

func (obj *FixedBuffer) WriteU16(value uint16, order binary.ByteOrder) {
	if !obj.CanWrite(2) {
		obj.Err = ErrBufferOverflow
		return
	}
	order.PutUint16(obj.buffer[obj.WritePos:obj.WritePos+2], value)
	obj.WritePos += 2
}

func (obj *FixedBuffer) WriteU32(value uint32, order binary.ByteOrder) {
	if !obj.CanWrite(4) {
		obj.Err = ErrBufferOverflow
		return
	}
	order.PutUint32(obj.buffer[obj.WritePos:obj.WritePos+4], value)
	obj.WritePos += 4
}

func (obj *FixedBuffer) WriteU64(value uint64, order binary.ByteOrder) {
	if !obj.CanWrite(8) {
		obj.Err = ErrBufferOverflow
		return
	}
	order.PutUint64(obj.buffer[obj.WritePos:obj.WritePos+8], value)
	obj.WritePos += 8
}

func (obj *FixedBuffer) WriteI64(value int64, order binary.ByteOrder) {
	obj.WriteU64(uint64(value), order)
}

func (obj *FixedBuffer) ReadU16(order binary.ByteOrder) (result uint16) {
	if !obj.CanRead(2) {
		obj.Err = ErrBufferOverflow
		return 0
	}

	result = order.Uint16(obj.buffer[obj.ReadPos:])
	obj.ReadPos += 2
	return result
}

func (obj *FixedBuffer) ReadU32(order binary.ByteOrder) (result uint32) {
	if !obj.CanRead(4) {
		obj.Err = ErrBufferOverflow
		return 0
	}
	result = order.Uint32(obj.buffer[obj.ReadPos:])
	obj.ReadPos += 4
	return result
}

func (obj *FixedBuffer) ReadU64(order binary.ByteOrder) (result uint64) {
	if !obj.CanRead(8) {
		obj.Err = ErrBufferOverflow
		return 0
	}
	result = order.Uint64(obj.buffer[obj.ReadPos:])
	obj.ReadPos += 8
	return result
}

func (obj *FixedBuffer) ReadF32(order binary.ByteOrder) float32 {
	bits := obj.ReadU32(order)
	result := math.Float32frombits(bits)
	obj.ReadPos += 4
	return result
}

func (obj *FixedBuffer) ReadF64(order binary.ByteOrder) float64 {
	bits := obj.ReadU64(order)
	result := math.Float64frombits(bits)
	obj.ReadPos += 8
	return result
}

func (obj *FixedBuffer) DumpDebug() {
	hexStr := hex.EncodeToString(obj.buffer[:obj.WritePos])
	fmt.Println(hexStr)
}

func (obj *FixedBuffer) StartReadMarker() {
	obj.readMarker = obj.ReadPos
}

func (obj *FixedBuffer) StartWriteMarker() {
	obj.writeMarker = obj.WritePos
}

func (obj *FixedBuffer) CalcReadCRC() uint16 {
	return crc16.ChecksumCCITTFalse(obj.buffer[obj.readMarker:obj.ReadPos])
}

func (obj *FixedBuffer) CalcWriteCRC() uint16 {
	return crc16.ChecksumCCITTFalse(obj.buffer[obj.writeMarker:obj.WritePos])
}

func (obj *FixedBuffer) WriteCRC16(order binary.ByteOrder) uint16 {
	value := obj.CalcWriteCRC()
	obj.WriteU16(value, order)

	return value
}

func (obj *FixedBuffer) AsReadSlice() []byte {
	return obj.buffer[:obj.ReadPos]
}

func (obj *FixedBuffer) AsWriteSlice() []byte {
	return obj.buffer[:obj.WritePos]
}

func (obj *FixedBuffer) WriteF32(speed float32, order binary.ByteOrder) {
	bits := math.Float32bits(speed)
	obj.WriteU32(bits, order)
}

func (obj *FixedBuffer) WriteF64(speed float64, order binary.ByteOrder) {
	bits := math.Float64bits(speed)
	obj.WriteU64(bits, order)
}

func (obj *FixedBuffer) AvailForRead() []byte {
	return obj.buffer[obj.ReadPos:]
}

func (obj *FixedBuffer) SeekRead(size int) {
	obj.ReadPos += size
}

func (obj *FixedBuffer) ReadBytes(size int) []byte {
	if !obj.CanRead(size) {
		obj.Err = ErrBufferOverflow
		return nil
	}

	res := obj.buffer[obj.ReadPos : obj.ReadPos+size]
	obj.ReadPos += size
	return res
}
