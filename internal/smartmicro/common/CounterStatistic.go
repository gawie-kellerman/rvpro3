package common

import (
	"encoding/binary"
	"time"

	"rvpro3/radarvision.com/utils"
)

type CounterStatistic struct {
	Id       int
	Count    uint64
	FirstOn  int64
	LastOn   int64
	ResetOn  int64
	IsActive bool
}

func (s *CounterStatistic) Add(count uint64, now time.Time) bool {
	if s.Count == 0 {
		s.Count = count
		s.FirstOn = now.Unix()
		s.LastOn = s.FirstOn
		return true
	}

	s.Count += count
	s.FirstOn = now.Unix()
	return false
}

func (s *CounterStatistic) WriteToFixedBuffer(writer *utils.FixedBuffer) {
	writer.WriteU16(uint16(s.Id), binary.LittleEndian)
	writer.WriteU64(s.Count, binary.LittleEndian)
	writer.WriteI64(s.FirstOn, binary.LittleEndian)
	writer.WriteI64(s.LastOn, binary.LittleEndian)
	writer.WriteI64(s.ResetOn, binary.LittleEndian)
	writer.WriteBool(s.IsActive)
}

var StatsHelper counterStatisticHelper

type counterStatisticHelper struct {
}

func (counterStatisticHelper) CountActives(stats []CounterStatistic) int {
	res := 0
	for _, stat := range stats {
		if stat.IsActive {
			res++
		}
	}
	return res
}
