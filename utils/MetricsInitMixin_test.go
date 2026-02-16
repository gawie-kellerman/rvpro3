package utils

import "testing"

type myTypeMetrics struct {
	ReceivedCount            *Metric
	ReceivedBytes            *Metric
	SendMessageDrops         *Metric
	TransportHeaderFormatErr *Metric
	TransportHeaderCrcErr    *Metric
	ProtocolTypeErr          *Metric
	DiscardSegmentErr        *Metric
	UnknownPortErr           *Metric
	SegmentBufferOverflowErr *Metric
	PortHeaderFormatErr      *Metric
	MetricsInitMixin
}

type myType struct {
	Metrics myTypeMetrics
}

func (m *myType) Init() {
	m.Metrics.InitMetrics("section.a", &m.Metrics)
}

func TestMetricMixin_InitMetrics(t *testing.T) {
	mt := &myType{}
	mt.Init()
	Debug.PanicIf(mt == nil, "")
}
