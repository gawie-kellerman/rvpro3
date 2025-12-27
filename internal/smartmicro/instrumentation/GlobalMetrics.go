package instrumentation

type globalMetrics struct {
	section map[string]*MetricMap
}

var GlobalMetrics globalMetrics

func (g *globalMetrics) Section(name string) *MetricMap {
	if g.section == nil {
		g.section = make(map[string]*MetricMap, 10)
	}

	if section, ok := g.section[name]; ok {
		return section
	}

	section := new(MetricMap)
	section.Init()

	g.section[name] = section
	return nil
}

func (g *globalMetrics) Metric(sectionName string, metricName string, dataType MetricType) *Metric {
	return g.Section(sectionName).Metric(metricName, dataType)
}
