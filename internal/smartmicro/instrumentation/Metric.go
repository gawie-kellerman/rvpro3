package instrumentation

import (
	"encoding/binary"
	"time"

	"rvpro3/radarvision.com/utils"
)

type MetricType uint8

const (
	MetricTypeUInt64 MetricType = iota
	MetricTypeTime
	MetricTypeDuration
)

type Metric struct {
	Id       int
	DataType MetricType
	Data     [8]byte
	FirstOn  int64
	LastOn   int64
	ResetOn  int64
	IsActive bool
}

func (s *Metric) AddCount(count uint64, now time.Time) bool {
	current := binary.LittleEndian.Uint64(s.Data[0:8])
	if current == 0 {
		s.IsActive = true
		s.DataType = MetricTypeUInt64
		s.FirstOn = now.Unix()
	}

	binary.LittleEndian.PutUint64(s.Data[0:8], current+count)
	s.LastOn = now.Unix()
	return current == 0
}

func (s *Metric) SetTime(tm time.Time) bool {
	current := binary.LittleEndian.Uint64(s.Data[0:8])

	if current == 0 {
		s.IsActive = true
		s.DataType = MetricTypeTime
		s.FirstOn = tm.UnixMilli()
	}

	binary.LittleEndian.PutUint64(s.Data[0:8], uint64(tm.UnixMilli()))
	s.LastOn = tm.UnixMilli()
	return current == 0
}

func (s *Metric) WriteToFixedBuffer(writer *utils.FixedBuffer) {
	writer.WriteU16(uint16(s.Id), binary.LittleEndian)
	writer.WriteU8(uint8(s.DataType))
	writer.WriteBytes(s.Data[0:8])
	writer.WriteI64(s.FirstOn, binary.LittleEndian)
	writer.WriteI64(s.LastOn, binary.LittleEndian)
	writer.WriteI64(s.ResetOn, binary.LittleEndian)
	writer.WriteBool(s.IsActive)
}

func (s *Metric) ReplaceMaxDuration(duration int64, now time.Time) {

	if !s.IsActive {
		s.IsActive = true
		s.DataType = MetricTypeDuration
		s.FirstOn = now.UnixMilli()
		s.LastOn = s.FirstOn
		binary.LittleEndian.PutUint64(s.Data[0:8], uint64(duration))
	} else {
		current := int64(binary.LittleEndian.Uint64(s.Data[0:8]))
		if current < duration {
			s.LastOn = now.UnixMilli()
			binary.LittleEndian.PutUint64(s.Data[0:8], uint64(duration))
		}
	}
}

func (s *Metric) ReplaceMinDuration(duration int64, now time.Time) {

	if !s.IsActive {
		s.IsActive = true
		s.DataType = MetricTypeDuration
		s.FirstOn = now.UnixMilli()
		s.LastOn = s.FirstOn
		binary.LittleEndian.PutUint64(s.Data[0:8], uint64(duration))
	} else {
		current := int64(binary.LittleEndian.Uint64(s.Data[0:8]))
		if current > duration {
			s.LastOn = now.UnixMilli()
			binary.LittleEndian.PutUint64(s.Data[0:8], uint64(duration))
		}
	}
}

var MetricsHelper metricsHelper

type metricsHelper struct {
}

func (metricsHelper) CountActives(stats []Metric) int {
	res := 0
	for _, stat := range stats {
		if stat.IsActive {
			res++
		}
	}
	return res
}
