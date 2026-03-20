package utils

import (
	"encoding/json"
	"fmt"
	"math"
)

type Uint64 uint64

func (u Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("0x%016x", u))
}

func (u Uint64) Combine(hi uint32, lo uint32) Uint64 {
	return Uint64(uint64(hi)<<32 | uint64(lo))
}

func (u Uint64) Split() (hi uint32, lo uint32) {
	hi = uint32(u >> 32)
	lo = uint32(uint64(u) & uint64(math.MaxUint32))
	return hi, lo
}

func (u Uint64) IsBit(index int) bool {
	return (u & (1 << uint(index))) != 0
}

func (u Uint64) SetBit(index int) Uint64 {
	return u | (1 << uint(index))
}

func (u Uint64) ClearBit(index int) Uint64 {
	return u &^ (1 << uint(index))
}
