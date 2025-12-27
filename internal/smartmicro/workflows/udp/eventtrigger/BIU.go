package eventtrigger

import (
	"math/bits"

	"rvpro3/radarvision.com/utils/bit"
)

const mask1 uint64 = 0xFFFF000000000000
const mask2 uint64 = 0x0000FFFF00000000
const mask3 uint64 = 0x00000000FFFF0000
const mask4 uint64 = 0x000000000000FFFF

type BIU uint64

func (b BIU) FromFlags(flags byte) BIU {
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

	return BIU(res)
}

func (b BIU) ToLE() uint64 {
	return bits.Reverse64(uint64(b))
}

func (b BIU) ToLEOld() uint64 {
	v := uint64(b)
	res := (v&0x000000000000FFFF)<<(64-16) |
		(v&0x00000000FFFF0000)<<(16) |
		(v&0x0000FFFF00000000)>>(16) |
		(v&0xFFFF000000000000)>>(48)
	return res
}
