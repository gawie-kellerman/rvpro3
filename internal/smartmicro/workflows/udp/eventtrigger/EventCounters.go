package eventtrigger

import "fmt"

const nofRelays = 8

// EventCounters only count when changing from lo(0) to hi(1)
// It only keeps count of the first 8 relay indexes that changes.  The indexes
// of the relays does not discriminate between hi (upper 32) and lo (lower 32)
// bits.
type EventCounters struct {
	Index        uint64
	RelayIndexes [nofRelays]uint32
	RelayCounts  [nofRelays]uint32
	Length       uint32
}

func (s *EventCounters) Dump() {
	for n := 0; n < int(s.Length); n++ {
		fmt.Printf("\n %d - index %d, count %d\n", n, s.RelayIndexes[n], s.RelayCounts[n])
	}
}

// Process to process a u32 of triggers
// oldValue is the previous trigger values
// newValue is the new trigger values
// set = 0 for low 32 bits, 1 for upper 32 bits
func (s *EventCounters) Process(oldValue uint32, newValue uint32, set int) {
	if s.Length < nofRelays {
		s.Parse(oldValue, newValue, set)
	}

	if s.Length > 0 {
		s.gather(oldValue, newValue, set)
	}
}

// Reset the counters and start counting with 0 length
func (s *EventCounters) Reset() {
	s.Length = 0
	s.Index = 0
}

func (s *EventCounters) GetLength() uint32 {
	return s.Length
}

func (s *EventCounters) Count(index int) (uint32, uint32) {
	return s.RelayIndexes[index] + 1, s.RelayCounts[index]
}

// Parse
// When lo=0, lower 32 bits
// When lo=1, upper 32 bits
func (s *EventCounters) Parse(oldValue uint32, newValue uint32, lo int) {
	offset := uint64(32 * lo)
	changed := ^oldValue & newValue

	for n := 0; n < 32 && s.Length < nofRelays; n++ {
		bitIndex := n + 32*lo
		dict := s.Index & (uint64(1) << (offset + uint64(n)))

		if changed&1 == 1 && dict == 0 {
			index := int(offset) + n
			s.RelayIndexes[s.Length] = uint32(index)
			s.RelayCounts[s.Length] = 0
			s.Length++

			s.Index |= uint64(1 << bitIndex)
		}
		changed >>= 1
	}
}

// gather counts the differences between the bits
func (s *EventCounters) gather(oldValue uint32, newValue uint32, set int) {
	lowerRange := 0
	upperRange := 31

	if set == 1 {
		lowerRange = 32
		upperRange = 63
	}

	for n := 0; n < int(s.Length); n++ {
		index := int(s.RelayIndexes[n])

		if index >= lowerRange && index <= upperRange {
			bitIndex := index - 32*set
			bit := uint32(1 << bitIndex)

			if oldValue&bit == 0 {
				if newValue&bit != 0 {
					s.RelayCounts[n] = min(s.RelayCounts[n]+1, 255)
				}
			}
		}
	}
}
