package utils

import (
	"encoding/binary"
	"encoding/json"
	"time"
)

type MetricType uint8

const (
	MetricTypeU64 MetricType = iota
	MetricTypeU32 MetricType = iota
	MetricTypeTime
	MetricTypeDuration
)

func (m MetricType) String() string {
	switch m {
	case MetricTypeU64:
		return "U64"
	case MetricTypeU32:
		return "U32"
	case MetricTypeTime:
		return "Time"
	case MetricTypeDuration:
		return "Duration"
	default:
		return "Unknown"
	}
}

type Metric struct {
	DataType MetricType
	Data     [8]byte
	FirstOn  int64
	LastOn   int64
	ResetOn  int64
	IsSet    bool
	Name     string
}

func (s *Metric) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Type":  s.DataType.String(),
		"First": time.UnixMilli(s.FirstOn).Format(DisplayDateTimeMS),
		"Last":  time.UnixMilli(s.LastOn).Format(DisplayDateTimeMS),
		"Reset": time.UnixMilli(s.ResetOn).Format(DisplayDateTimeMS),
		"Value": s.GetValue(),
	})
}

func (s *Metric) AddCount(count uint64, now time.Time) bool {
	wasActive := s.IsSet
	current := binary.LittleEndian.Uint64(s.Data[0:8])

	if !s.IsSet {
		s.IsSet = true
		s.DataType = MetricTypeU64
		s.FirstOn = now.UnixMilli()
	}

	binary.LittleEndian.PutUint64(s.Data[0:8], current+count)
	s.LastOn = now.UnixMilli()
	return wasActive
}

func (s *Metric) SetTime(tm time.Time) bool {
	wasActive := s.IsSet

	if !s.IsSet {
		s.IsSet = true
		s.DataType = MetricTypeTime
		s.FirstOn = tm.UnixMilli()
	}

	binary.LittleEndian.PutUint64(s.Data[0:8], uint64(tm.UnixMilli()))
	s.LastOn = tm.UnixMilli()
	return wasActive
}

func (s *Metric) GetTime() time.Time {
	iValue := int64(binary.LittleEndian.Uint64(s.Data[0:8]))
	return time.UnixMilli(iValue)
}

func (s *Metric) WriteToFixedBuffer(writer *FixedBuffer) {
	writer.WritePascal(s.Name)
	writer.WriteU8(uint8(s.DataType))
	writer.WriteBytes(s.Data[0:8])
	writer.WriteI64(s.FirstOn, binary.LittleEndian)
	writer.WriteI64(s.LastOn, binary.LittleEndian)
	writer.WriteI64(s.ResetOn, binary.LittleEndian)
	writer.WriteBool(s.IsSet)
}

func (s *Metric) ReplaceMaxDuration(duration int64, now time.Time) {
	if !s.IsSet {
		s.IsSet = true
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
	if !s.IsSet {
		s.IsSet = true
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

func (s *Metric) SetU32(value uint32, now time.Time) {
	if !s.IsSet {
		s.IsSet = true
		s.FirstOn = now.UnixMilli()
		s.DataType = MetricTypeU32
	}
	s.LastOn = now.UnixMilli()
	binary.LittleEndian.PutUint32(s.Data[0:4], value)
}

func (s *Metric) GetU32() uint32 {
	return binary.LittleEndian.Uint32(s.Data[0:4])
}

func (s *Metric) SetU16(value uint16, now time.Time) {
	if !s.IsSet {
		s.IsSet = true
		s.FirstOn = now.UnixMilli()
		s.DataType = MetricTypeU32
	}
	s.LastOn = now.UnixMilli()
	binary.LittleEndian.PutUint16(s.Data[0:4], value)
}

func (s *Metric) GetValue() interface{} {
	switch s.DataType {
	case MetricTypeU64:
		return s.GetU64()
	case MetricTypeU32:
		return s.GetU32()
	case MetricTypeTime:
		return s.GetU64()
	case MetricTypeDuration:
		return s.GetU64()
	default:
		return nil
	}
}

func (s *Metric) GetU64() interface{} {
	return binary.LittleEndian.Uint64(s.Data[0:8])
}

func (s *Metric) Inc(now time.Time) {
	current := binary.LittleEndian.Uint64(s.Data[0:8])
	if !s.IsSet {
		s.IsSet = true
		s.DataType = MetricTypeU64
		s.FirstOn = now.UnixMilli()
	}

	binary.LittleEndian.PutUint64(s.Data[0:8], current+1)
	s.LastOn = now.UnixMilli()
}

func (s *Metric) Add(count int, now time.Time) {
	s.AddCount(uint64(count), now)
}

func (s *Metric) IncRes(now time.Time) uint64 {
	current := binary.LittleEndian.Uint64(s.Data[0:8])
	if !s.IsSet {
		s.IsSet = true
		s.DataType = MetricTypeU64
		s.FirstOn = now.UnixMilli()
	}

	binary.LittleEndian.PutUint64(s.Data[0:8], current+1)
	s.LastOn = now.UnixMilli()

	return current + 1
}
