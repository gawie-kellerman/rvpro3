package utils

import (
	"maps"
	"regexp"
	"strings"
	"sync"
)

// globalMetrics only safeguard against
// the root section map.  The use pattern is to register the metric to the map
// but keep a local reference in order to (a) avoid the mutex and (b) avoid the map lookups.
// This is advisable as there is generally very little risk in this approach
// where the data is concerned, especially as the running counters gives
// meaning and corrects itself over time.
// It does not need to be accurate 12 times a second.  We favor efficiency
// over absolute effectiveness, as it is needed.
type globalMetrics struct {
	section map[string]*Metrics
	mutex   sync.RWMutex
}

func (g *globalMetrics) Init() {
	g.section = make(map[string]*Metrics, 10)
}

func (g *globalMetrics) Section(name string) *Metrics {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if section, ok := g.section[name]; ok {
		return section
	}

	section := new(Metrics)
	section.Init(name)

	g.section[name] = section
	return section
}

func (g *globalMetrics) Metric(sectionName string, metricName string, dataType MetricType) *Metric {
	return g.Section(sectionName).GetOrPut(metricName, dataType)
}

func (g *globalMetrics) StartsWith(sectionName string) []string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	res := make([]string, 0, 4)

	for name := range maps.Keys(g.section) {
		if strings.HasPrefix(name, sectionName) {
			res = append(res, name)
		}
	}
	return res
}

func (g *globalMetrics) SectionsStartWith(sectionName string) []*Metrics {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	names := g.StartsWith(sectionName)
	res := make([]*Metrics, 0, len(names))

	for _, name := range names {
		res = append(res, g.Section(name))
	}
	return res
}

func (g *globalMetrics) Names() []string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	res := make([]string, 0, len(g.section))
	for name := range maps.Keys(g.section) {
		res = append(res, name)
	}
	return res
}

func (g *globalMetrics) FindOrNil(name string) *Metrics {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if metric, ok := g.section[name]; ok {
		return metric
	}
	return nil
}

func (g *globalMetrics) MergeRegEx(res map[string]*Metrics, regEx string) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	for metricName, metric := range g.section {
		if matched, _ := regexp.MatchString(regEx, metricName); matched {
			res[metricName] = metric
		}
	}
}
