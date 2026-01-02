package bit

type U64Bits uint64

// ForEachU64Bit loops with filtering over the bits of a uint64.
// It assumes that chances are that only a few bits are used and that looping over
// the whole array will result in very few true callbacks.
func (source U64Bits) ForEachU64Bit(callback func(index int, isSet bool)) {
	// loop over 8 bytes of source
	for n := 0; n < 8; n++ {
		// Check if a byte has bits set
		if source.HasBits(n) {
			source.For(n*8, ((n+1)*8)-1, callback)
		}
	}
}

func (source U64Bits) HasBits(block int) bool {
	mask := uint64(0xFF << uint(block*8))

	return uint64(source)&mask != 0
}

// For loops over a part of a uint64 value.
func (source U64Bits) For(start, end int, callback func(index int, isSet bool)) {
	bit := uint64(1) << uint(start)

	if start > end {
		for n := start; n >= end; n-- {
			callback(n, uint64(source)&bit != 0)
			bit >>= 1
		}
	} else {
		for n := start; n <= end; n++ {
			callback(n, uint64(source)&bit != 0)
			bit <<= 1
		}
	}
}
