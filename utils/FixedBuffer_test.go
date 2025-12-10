package utils

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixedBuffer_Writing(t *testing.T) {
	fixed := FixedBuffer{
		Buffer: make([]byte, 1024),
	}

	fixed.WriteU8(1)
	fixed.WriteU8(2)
	fixed.WriteU16(3, binary.BigEndian)
	fixed.WriteCRC16(binary.BigEndian)

	fixed.DumpDebug()
}

func TestFixedBuffer_Reading(t *testing.T) {
	fixed := FixedBuffer{
		Buffer: make([]byte, 1024),
	}
	fixed.WriteU8(1)
	fixed.WriteU8(2)
	fixed.WriteU16(3, binary.BigEndian)

	assert.Equal(t, uint8(1), fixed.ReadU8())
	assert.Equal(t, uint8(2), fixed.ReadU8())
	assert.Equal(t, uint16(3), fixed.ReadU16(binary.BigEndian))
}
