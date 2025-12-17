package instrumentation

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

type RadarMetricType int

const (
	radarMetricHead RadarMetricType = iota + 200
	RmtTotalMessagesDropped
	RmtTotalMessagesProcessed
	RmtTotalBytesProcessed
	RmtObjectListProcessed
	RmtObjectListDropped
	FmtObjectListTotalTime
	RmtObjectListMinTime
	RmtObjectListMaxTime
	RmtStatisticsProcessed
	RmtStatisticsDropped
	RmtStatisticsTotalTime
	RmtStatisticsMinTime
	RmtStatisticsMaxTime
	RmtTriggerProcessed
	RmtTriggerDropped
	RmtTriggerTotalTime
	RmtTriggerMinTime
	RmtTriggerMaxTime
	RmtPVRProcessed
	RmtPVRDropped
	RmtPVRTotalTime
	RmtPVRMinTime
	RmtPVRMaxTime
	RmtInstructionProcessed
	RmtInstructionDropped
	RmtInstructionTotalTime
	RmtInstructionMinTime
	RmtInstructionMaxTime
	RmtDiagnosticProcessed
	RmtDiagnosticDropped
	RmtDiagnosticTotalTime
	RmtDiagnosticMinTime
	RmtDiagnosticMaxTime
	RmtTransportHeaderFormatErr
	RmtHeaderFormatErr
	RmtTransportHeaderCRCErr
	RmtProtocolTypeErr
	RmtDiscardedSegment
	RmtUnknownPortIdentifier
	RmtUnknownDropped
	RmtSegmentationBufferOverflow
	RmtSegmentationDropped
	radarMetricTail
)

type RadarMetrics struct {
	RadarIP utils.IP4
	Metrics
}

type globalRadarMetrics struct {
	Radar [4]RadarMetrics
}

func (g *globalRadarMetrics) ByIndex(index int) *RadarMetrics {
	return &g.Radar[index]
}

var GlobalRadarMetrics globalRadarMetrics

func init() {
	GlobalRadarMetrics.Radar[0].Metrics.SetLength(int(radarMetricHead), int(radarMetricTail))
	GlobalRadarMetrics.Radar[1].Metrics.SetLength(int(radarMetricHead), int(radarMetricTail))
	GlobalRadarMetrics.Radar[2].Metrics.SetLength(int(radarMetricHead), int(radarMetricTail))
	GlobalRadarMetrics.Radar[3].Metrics.SetLength(int(radarMetricHead), int(radarMetricTail))
}
