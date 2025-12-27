package uartsdlc

import (
	"encoding/binary"

	"github.com/howeyc/crc16"
	"github.com/pkg/errors"
)

const startMarker byte = 0x02
const endMarker byte = 0x03
const escapeMarker byte = 0x7D
const escapeMask byte = 0x20

var ErrSDLCTransformTargetOverflow = errors.New("SDLC transform target rawData overflow")
var ErrSDLCCRCOverflow = errors.New("SDLC CRC16 overflow")
var ErrSDLCInvalidStartMarker = errors.New("SDLC Invalid start marker")
var ErrSDLCInvalidEndMarker = errors.New("SDLC Invalid end marker")
var ErrSDLCInvalidBuffer = errors.New("SDLC Invalid rawData")
var ErrSDLCCRCCheck = errors.New("SDLC CRC Check")

type codec struct {
}

// Codec formats and changes existing buffers.
// It never escapes temporary buffers to the heap.
// When encoding create buffers with sufficient additional space
// When decoding the rawData can only shrink
var Codec codec

// encodeData takes a byte array and preprocesses it to escape all control characters
// The control characters include the startMarker, endMarker, and escapeMarker.
// Note that since these are all escaped, that legitimate start and end should not be
// part of the source parameter.  You have to call appendCRC before calling this method
func (codec) encodeData(source []byte) ([]byte, error) {
	var target [256]byte
	var targetIdx int

	for sourceIdx := 0; sourceIdx < len(source); sourceIdx++ {
		sourceByte := source[sourceIdx]
		if Codec.isDelimiter(sourceByte) {
			target[targetIdx] = escapeMarker
			targetIdx++
			target[targetIdx] = sourceByte ^ escapeMask
		} else {
			target[targetIdx] = sourceByte
		}
		targetIdx++

		if targetIdx >= len(target) {
			return nil, ErrSDLCTransformTargetOverflow
		}
	}

	source = source[:targetIdx]
	copy(source, target[:targetIdx])
	return source, nil
}

// appendCRC appends a 16-bit crc.  The CRC still needs to be encoded
func (codec) appendCRC(data []byte) ([]byte, error) {
	if len(data)+2 >= cap(data) {
		return nil, ErrSDLCCRCOverflow
	}

	res := data[:len(data)+2]

	crc := crc16.ChecksumCCITTFalse(data)
	binary.BigEndian.PutUint16(res[len(data):], crc)
	return res, nil
}

// Decode an SDLC rawData in place.  The result is a slice of the
// original rawData; the slice is equal or smaller than the length of the
// rawData.
func (codec) Decode(buffer []byte) (res []byte, err error) {
	if len(buffer) <= 5 {
		return nil, ErrSDLCInvalidBuffer
	}

	if buffer[0] != startMarker {
		return nil, ErrSDLCInvalidStartMarker
	}

	if buffer[len(buffer)-1] != endMarker {
		return nil, ErrSDLCInvalidEndMarker
	}

	// The rawData is between the start and end markers
	data := buffer[1 : len(buffer)-1]
	data = Codec.decodeData(data)

	// don't perform the crc16 on the crc16
	crcCheck := crc16.ChecksumCCITTFalse(data[:len(data)-2])
	crc := binary.BigEndian.Uint16(data[len(data)-2:])

	if crc != crcCheck {
		return nil, ErrSDLCCRCCheck
	}

	data = append(data, endMarker)
	return buffer[:len(data)+1], nil
}

func (codec) DecodeInto(source []byte, target []byte) (res []byte, err error) {
	if len(source) <= 5 {
		return nil, ErrSDLCInvalidBuffer
	}

	if source[0] != startMarker {
		return nil, ErrSDLCInvalidStartMarker
	}

	if source[len(source)-1] != endMarker {
		return nil, ErrSDLCInvalidEndMarker
	}

	targetIdx := 0
	escaped := false

	for sourceIdx := 0; sourceIdx < len(source); sourceIdx++ {
		if targetIdx >= len(target) {
			return nil, ErrSDLCTransformTargetOverflow
		}
		sourceByte := source[sourceIdx]

		if escaped {
			target[targetIdx] = sourceByte ^ escapeMask
			targetIdx++
			escaped = false
		} else if sourceByte == escapeMarker {
			escaped = true
		} else {
			target[targetIdx] = sourceByte
			targetIdx++
		}
	}

	res = target[:targetIdx]

	if err = Codec.CheckCRC(res); err != nil {
		return nil, err
	}

	return res, nil
}

// decodeData takes a byte array as SDLC UART, parse and convert the escaped data
// to data.  It does this inplace as the results is always equal to or smaller than
// the escaped data; meaning the rawData will always fit.
func (codec) decodeData(source []byte) []byte {
	targetIdx := 0
	escaped := false

	for sourceIdx := 0; sourceIdx < len(source); sourceIdx++ {
		sourceByte := source[sourceIdx]
		if escaped {
			source[targetIdx] = sourceByte ^ escapeMask
			targetIdx++
			escaped = false
		} else if sourceByte == escapeMarker {
			escaped = true
		} else {
			source[targetIdx] = sourceByte
			targetIdx++
		}
	}

	return source[:targetIdx]
}

// Encode must be called with the data in the rawData starting at position 1.
// This is to allow adding the start marker at position 0.  The rawData must also
// have enough capacity for the escaped characters and +2 for the start and end markers
func (codec) Encode(buffer []byte) (res []byte, err error) {
	data := buffer[1:]

	if data, err = Codec.appendCRC(data); err != nil {
		return nil, err
	}

	if data, err = Codec.encodeData(data); err != nil {
		return nil, err
	}

	res = buffer[:1+len(data)+1]
	res[0] = startMarker
	res[len(res)-1] = endMarker
	return res, nil
}

func (codec) isDelimiter(sourceByte byte) bool {
	return sourceByte == startMarker || sourceByte == endMarker || sourceByte == escapeMarker
}

func (codec) CalcCRC(buffer []byte) uint16 {
	crcSlice := buffer[1 : len(buffer)-3]
	return crc16.ChecksumCCITTFalse(crcSlice)
}

func (codec) GetCRC(buffer []byte) uint16 {
	off := len(buffer) - 3
	return binary.BigEndian.Uint16(buffer[off : off+2])
}

func (codec) CheckCRC(buffer []byte) error {
	check := Codec.CalcCRC(buffer)
	crc := Codec.GetCRC(buffer)

	if check != crc {
		return ErrSDLCCRCCheck
	}
	return nil
}

// GetDataLen assumes that slice is a raw-data (unescaped) rawData.
func (c codec) GetDataLen(slice []byte) int {
	// 5 => 2 for header bytes, 3 for trailer bytes
	return len(slice) - 5
}
