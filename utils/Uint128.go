package utils

import (
	"fmt"
	"strings"

	"rvpro3/radarvision.com/utils/bit"
)

type Uint128 struct {
	Hi uint64
	Lo uint64
}

type callback func(int, bool)

func (u Uint128) Byte(index int) byte {
	source := u.Lo

	if index > 7 {
		source = u.Hi
		index -= 8
	}
	index *= 8

	return byte((source >> index) & 0xFF)
}

func (u Uint128) For(start, end int, cb callback) {
	if start <= end {
		for i := start; i < end; i++ {
			cb(i, u.IsBit(i))
		}
	} else {
		for i := start; i >= end; i-- {
			cb(i, u.IsBit(i))
		}
	}
}

func (u Uint128) IsBit(i int) bool {
	if i >= 64 {
		return bit.IsSet(u.Hi, i-64)
	} else {
		return bit.IsSet(u.Lo, i)
	}
}

func (u Uint128) SetBit(index int, value bool) Uint128 {
	res := u
	if index >= 64 {
		res.Hi = bit.SetOnOrOff(u.Hi, index-64, value)
	} else {
		res.Lo = bit.SetOnOrOff(u.Lo, index, value)
	}
	return res
}

func (u Uint128) And(other Uint128) Uint128 {
	return Uint128{
		Hi: u.Hi & other.Hi,
		Lo: u.Lo & other.Lo,
	}
}

func (u Uint128) Or(other Uint128) Uint128 {
	return Uint128{
		Hi: u.Hi | other.Hi,
		Lo: u.Lo | other.Lo,
	}
}

func (u Uint128) OrLo(other uint64) Uint128 {
	return Uint128{
		Hi: u.Hi,
		Lo: u.Lo | other,
	}
}

func (u Uint128) Hex() string {
	return fmt.Sprintf("%016x%016x", u.Hi, u.Lo)
}

func (u Uint128) AreBitsEqual(start int, size int, other Uint128) bool {
	for n := start; n < start+size; n++ {
		if u.IsBit(n) != other.IsBit(n) {
			return false
		}
	}
	return true
}

func (u Uint128) Equals(other Uint128) bool {
	return u.Lo == other.Lo && u.Hi == other.Hi
}

func (u Uint128) ToString(count int, msbFirst bool) string {
	bld := strings.Builder{}
	if msbFirst {
		for n := 0; n < count; n++ {
			if u.IsBit(n) {
				bld.WriteString("1")
			} else {
				bld.WriteString("0")
			}
		}
	} else {
		for n := count; n >= 0; n-- {
			if u.IsBit(n) {
				bld.WriteString("1")
			} else {
				bld.WriteString("0")
			}
		}
	}
	return bld.String()
}

func (u Uint128) String() string {
	return fmt.Sprintf("%016x%016x", u.Hi, u.Lo)
}

func (u Uint128) Negate() Uint128 {
	return Uint128{
		Hi: ^u.Lo,
		Lo: ^u.Lo,
	}
}

func (u Uint128) FromString(flags string, offValue uint8) {
	for i := 0; i < len(flags); i++ {
		u.SetBit(i, flags[i] != offValue)
	}
}

func (u Uint128) Equals64(hi uint64, lo uint64) bool {
	return u.Hi == hi && u.Lo == lo
}
