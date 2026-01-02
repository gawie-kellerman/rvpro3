package uartsdlc

import (
	"strings"

	"rvpro3/radarvision.com/utils/bit"
)

const mask1 uint64 = 0xFFFF000000000000
const mask2 uint64 = 0x0000FFFF00000000
const mask3 uint64 = 0x00000000FFFF0000
const mask4 uint64 = 0x000000000000FFFF

type BIUMask uint64
type BIUFlags uint8

func (flags BIUFlags) ToBIUMask() BIUMask {
	var res uint64

	if bit.IsSet(flags, 0) {
		res |= mask1
	}

	if bit.IsSet(flags, 1) {
		res |= mask2
	}

	if bit.IsSet(flags, 2) {
		res |= mask3
	}

	if bit.IsSet(flags, 3) {
		res |= mask4
	}

	return BIUMask(res)
}

func (flags BIUFlags) String() string {
	str := strings.Builder{}
	str.Grow(64)

	for i := 0; i < 4; i++ {
		if bit.IsSet(flags, i) {
			str.WriteString("1111111111111111")
		} else {
			str.WriteString("0000000000000000")
		}
	}
	return str.String()
}
