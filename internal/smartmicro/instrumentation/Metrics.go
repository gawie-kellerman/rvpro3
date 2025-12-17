package instrumentation

import (
	"time"
)

type Metrics struct {
	Name string
	Data []Metric
	Head int
	Tail int
}

func (s *Metrics) SetLength(head int, tail int) {
	s.Head = head
	s.Tail = tail
	s.Data = make([]Metric, 0, tail-head-1)
}

func (s *Metrics) GetRel(key int) *Metric {
	index := key - s.Head - 1

	return &s.Data[index]
}

// AddCount return true if the count was zero before setting
func (s *Metrics) AddCount(rel int, count uint64, time time.Time) bool {
	metric := s.GetRel(rel)
	return metric.AddCount(count, time)
}

func (s *Metrics) SetTime(rel int, now time.Time) bool {
	metric := s.GetRel(rel)
	return metric.SetTime(now)
}

func (s *Metrics) ReplaceMaxDuration(rel int, duration int64, now time.Time) {
	metric := s.GetRel(rel)
	metric.ReplaceMaxDuration(duration, now)
}
