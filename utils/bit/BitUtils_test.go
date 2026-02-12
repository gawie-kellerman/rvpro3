package bit

import (
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverse64(t *testing.T) {
	one := uint8(1)
	assert.Equal(t, uint8(128), bits.Reverse8(one))
}

func TestString(t *testing.T) {
	assert.Equal(t, "00000011", AsString(byte(3)))
	assert.Equal(t, "00000100", AsString(byte(4)))
}

func TestClear(t *testing.T) {
	//  0  1  2  3  4
	//  1  2  4  8 16
	assert.Equal(t, uint8(127), Clear(byte(255), 7))
	assert.Equal(t, uint8(0), Clear(byte(1), 0))
	assert.Equal(t, uint8(14), Clear(byte(15), 0))
	assert.Equal(t, uint8(7), Clear(byte(15), 3))
	assert.Equal(t, uint8(15), Clear(byte(15), 4))
}

func TestSet(t *testing.T) {
	assert.Equal(t, uint8(128), Set(byte(0), 7))
	assert.Equal(t, uint8(129), Set(byte(1), 7))
}

func TestIsSet(t *testing.T) {
	assert.Equal(t, true, IsSet(byte(128), 7))
	assert.Equal(t, true, IsSet(byte(1), 0))
	assert.Equal(t, false, IsSet(byte(1), 1))
}
