package instrumentation

import (
	"encoding/json"
	"time"
)

const MetricStartForUDP = 1000
const MetricStartForRadar = 2000
const MetricStartForSDLC = 3000

type Metrics struct {
	Name string
	Data []Metric
	Head int
	Tail int
}

type MetricMap struct {
	Data map[string]*Metric
}

func (m *MetricMap) Init() {
	m.Data = make(map[string]*Metric, 10)
}

func (m *MetricMap) Metric(name string, dataType MetricType) *Metric {
	if metric, ok := m.Data[name]; ok {
		return metric
	}

	metric := new(Metric)
	metric.DataType = dataType
	metric.IsActive = true
	metric.Name = name
	m.Data[name] = metric

	return metric
}

func (s *Metrics) MarshalJSON() ([]byte, error) {
	dataPointers := make([]*Metric, 0, len(s.Data))

	for _, data := range s.Data {
		if data.IsActive {
			dataPointers = append(dataPointers, &data)
		}
	}
	return json.Marshal(dataPointers)
}

func (s *Metrics) SetLength(head int, tail int) {
	s.Head = head
	s.Tail = tail
	s.Data = make([]Metric, tail-head-1)
}

func (s *Metrics) GetRel(key int) *Metric {
	index := key - s.Head - 1

	res := &s.Data[index]
	res.Id = key
	return res
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
