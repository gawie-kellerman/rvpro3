package utils

import (
	"fmt"
	"strings"
	"time"
)

type MetricType uint8

const (
	MtI64 MetricType = iota
	MtMilliTime
	MtMilliDuration
	MtBool
	MtUnknown
)

var metricTypeAbbr = [...]string{"I64", "MT", "MD", "YN", "UNK"}

func (m *MetricType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", metricTypeAbbr[*m])), nil
}

func (m *MetricType) UnmarshalJSON(data []byte) error {
	src := strings.Trim(string(data), "\"")
	for index, name := range metricTypeAbbr {
		if src == name {
			*m = MetricType(index)
		}
	}

	*m = MtUnknown
	return nil
}

type Metric struct {
	Name    string `json:"-"`
	FirstOn int64  `json:"FirstOn,omitempty"`
	LastOn  int64  `json:"LastOn,omitempty"`
	ResetOn int64  `json:"ResetOn,omitempty"`
	IsSet   bool   `json:"-"`
	Type    MetricType
	Value   int64
}

func (s *Metric) Inc(value int64) {
	now := time.Now().UnixMilli()

	if !s.IsSet {
		s.IsSet = true
		s.FirstOn = now
	}

	s.Value = s.Value + value
	s.LastOn = now
}

func (s *Metric) IncAt(value int64, tm time.Time) bool {
	now := tm.UnixMilli()

	if !s.IsSet {
		s.IsSet = true
		s.FirstOn = now
		s.Value = s.Value + value
		s.LastOn = now
		return true
	}

	s.Value = s.Value + value
	s.LastOn = now
	return false
}

func (s *Metric) Dec(value int64) {
	now := time.Now().UnixMilli()

	if !s.IsSet {
		s.IsSet = true
		s.FirstOn = now
	}

	s.Value = s.Value - value
	s.LastOn = now
}

func (s *Metric) Set(value int64) {
	now := time.Now().UnixMilli()

	if !s.IsSet {
		s.IsSet = true
		s.FirstOn = now
	}

	s.Value = s.Value - value
	s.LastOn = now
}

func (s *Metric) SetBool(value bool) {
	if value {
		s.Set(1)
	} else {
		s.Set(0)
	}
}

func (s *Metric) GetBool() bool {
	return s.Value&1 == 1
}

func (s *Metric) SetTime() {
	now := time.Now().UnixMilli()

	if !s.IsSet {
		s.IsSet = true
		s.Type = MtI64
		s.FirstOn = now
	}

	s.Value = now
	s.LastOn = now
}

func (s *Metric) GetTime() time.Time {
	return time.UnixMilli(s.Value)
}

//func (s *Metric) WriteToFixedBuffer(writer *FixedBuffer) {
//	writer.WritePascal(s.Name)
//	writer.WriteU8(uint8(s.Type))
//	writer.WriteBytes(s.Value[0:8])
//	writer.WriteI64(s.FirstOn, binary.LittleEndian)
//	writer.WriteI64(s.LastOn, binary.LittleEndian)
//	writer.WriteI64(s.ResetOn, binary.LittleEndian)
//	writer.WriteBool(s.IsSet)
//}

func (s *Metric) SetIfMore(value int64) {
	if !s.IsSet {
		now := time.Now().UnixMilli()
		s.IsSet = true
		s.FirstOn = now
		s.LastOn = now
		s.Value = value
	} else {
		if value > s.Value {
			now := time.Now().UnixMilli()
			s.LastOn = now
			s.Value = value
		}
	}
}

func (s *Metric) SetIfMoreAt(value int64, tm time.Time) {
	if !s.IsSet {
		now := tm.UnixMilli()
		s.IsSet = true
		s.FirstOn = now
		s.LastOn = now
		s.Value = value
	} else {
		if value > s.Value {
			now := tm.UnixMilli()
			s.LastOn = now
			s.Value = value
		}
	}
}

func (s *Metric) SetIfLess(value int64) {
	if !s.IsSet {
		now := time.Now().UnixMilli()
		s.IsSet = true
		s.FirstOn = now
		s.LastOn = now
		s.Value = value
	} else {
		if value < s.Value {
			now := time.Now().UnixMilli()
			s.LastOn = now
			s.Value = value
		}
	}
}

func (s *Metric) SetIfLessAt(value int64, tm time.Time) {
	if !s.IsSet {
		now := tm.UnixMilli()
		s.IsSet = true
		s.FirstOn = now
		s.LastOn = now
		s.Value = value
	} else {
		if value < s.Value {
			now := tm.UnixMilli()
			s.LastOn = now
			s.Value = value
		}
	}
}
