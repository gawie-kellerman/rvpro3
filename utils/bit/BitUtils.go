package bit

import (
	"encoding/binary"
	"strings"

	"golang.org/x/exp/constraints"
)

func CombineU32(hi uint32, lo uint32) uint64 {
	res := uint64(hi) << 32
	res |= uint64(lo)
	return res
}

func IsSet[T constraints.Integer](value T, bit int) bool {
	flags := T(1 << bit)
	return value&flags == flags
}

func Set[T constraints.Integer](value T, bit int) T {
	flags := T(1 << bit)
	value |= flags
	return value
}

func Clear[T constraints.Integer](value T, bit int) T {
	flags := T(1 << bit)
	value &= ^flags
	return value
}

func SetOnOrOff[T constraints.Integer](value T, bit int, on bool) T {
	if on {
		return Set(value, bit)
	}
	return Clear(value, bit)
}

func ToBuilder[T constraints.Integer](bld *strings.Builder, value T) {
	bits := binary.Size(value) * 8
	bld.Grow(bits)

	mask := T(1) << (bits - 1)

	for n := bits; n > 0; n-- {
		if value&mask != 0 {
			bld.WriteRune('1')
		} else {
			bld.WriteRune('0')
		}
		mask >>= 1
	}
}

func AsString[T constraints.Integer](value T) string {
	bld := strings.Builder{}
	ToBuilder(&bld, value)
	return bld.String()
}

func ForLSB[T constraints.Integer](value T, callback func(int, bool)) {
	check := T(1)
	bits := binary.Size(value) * 8

	for n := 0; n < bits; n++ {
		callback(n, check&value == check)
		check <<= 1
	}
}
