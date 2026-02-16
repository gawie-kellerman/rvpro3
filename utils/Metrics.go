package utils

type Metrics struct {
	Name   string `json:"-"`
	Metric map[string]*Metric
}

func (m *Metrics) Init(sectionName string) {
	m.Metric = make(map[string]*Metric, 10)
	m.Name = sectionName
}

func (m *Metrics) GetOrPut(name string, dataType MetricType) *Metric {
	if m.Metric == nil {
		m.Init(name)
	}

	if metric, ok := m.Metric[name]; ok {
		return metric
	}

	metric := new(Metric)
	metric.Type = dataType
	metric.IsSet = false
	metric.Name = name
	m.Metric[name] = metric

	return metric
}

func (m *Metrics) Get(name string) *Metric {
	return m.Metric[name]
}

//func (s *Metrics) MarshalJSON() ([]byte, error) {
//	dataPointers := make([]*GetOrPut, 0, len(s.Metric))
//
//	for _, data := range s.Metric {
//		if data.IsSet {
//			dataPointers = append(dataPointers, &data)
//		}
//	}
//	return json.Marshal(dataPointers)
//}
