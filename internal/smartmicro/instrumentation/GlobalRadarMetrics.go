package instrumentation

import (
	"encoding/json"

	"rvpro3/radarvision.com/utils"
)

type RadarMetricType int

const (
	radarMetricHead RadarMetricType = iota + MetricStartForRadar
	RmtTotalMessagesDropped
	RmtTotalMessagesProcessed
	RmtTotalBytesProcessed
	RmtObjectListProcessed
	RmtObjectListDropped
	RmtObjectListTotalTime
	RmtObjectListMinTime
	RmtObjectListMaxTime
	RmtStatisticsProcessed
	RmtStatisticsDropped
	RmtStatisticsTotalTime
	RmtStatisticsMinTime //
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

func init() {
	gm := GlobalMetricNames()
	gm[int(RmtTotalMessagesDropped)] = "Total Messages Dropped"
	gm[int(RmtTotalMessagesProcessed)] = "Total Messages Processed"
	gm[int(RmtTotalBytesProcessed)] = "Total Bytes Processed"
	gm[int(RmtObjectListProcessed)] = "Object List Processed"
	gm[int(RmtObjectListDropped)] = "Object List Dropped"
	gm[int(RmtObjectListTotalTime)] = "Object List Total Time"
	gm[int(RmtObjectListMinTime)] = "Object List Min Time"
	gm[int(RmtObjectListMaxTime)] = "Object List Max Time"
	gm[int(RmtStatisticsProcessed)] = "Statistics Processed"
	gm[int(RmtStatisticsDropped)] = "Statistics Dropped"
	gm[int(RmtStatisticsTotalTime)] = "Statistics Total Time"
	gm[int(RmtStatisticsMinTime)] = "Statistics Min Time"
	gm[int(RmtStatisticsMaxTime)] = "Statistics Max Time"
	gm[int(RmtTriggerProcessed)] = "Trigger Processed"
	gm[int(RmtTriggerDropped)] = "Trigger Dropped"
	gm[int(RmtTriggerTotalTime)] = "Trigger Total Time"
	gm[int(RmtTriggerMinTime)] = "Trigger Min Time"
	gm[int(RmtTriggerMaxTime)] = "Trigger Max Time"
	gm[int(RmtTriggerTotalTime)] = "Trigger Total Time"
	gm[int(RmtTriggerMinTime)] = "Trigger Min Time"
	gm[int(RmtTriggerMaxTime)] = "Trigger Max Time"
	gm[int(RmtPVRDropped)] = "PVR Dropped"
	gm[int(RmtPVRProcessed)] = "PVR Processed"
	gm[int(RmtPVRMaxTime)] = "PVR Max Time"
	gm[int(RmtPVRMinTime)] = "PVR Min Time"
	gm[int(RmtPVRTotalTime)] = "PVR Total Time"
	gm[int(RmtInstructionMinTime)] = "Instruction Min Time"
	gm[int(RmtInstructionMaxTime)] = "Instruction Max Time"
	gm[int(RmtInstructionDropped)] = "Instruction Dropped"
	gm[int(RmtInstructionProcessed)] = "Instruction Processed"
	gm[int(RmtInstructionTotalTime)] = "Instruction Total Time"
	gm[int(RmtDiagnosticProcessed)] = "Diagnostic Processed"
	gm[int(RmtDiagnosticDropped)] = "Diagnostic Dropped"
	gm[int(RmtDiagnosticTotalTime)] = "Diagnostic Total Time"
	gm[int(RmtDiagnosticMinTime)] = "Diagnostic Min Time"
	gm[int(RmtDiagnosticMaxTime)] = "Diagnostic Max Time"
	gm[int(RmtTransportHeaderFormatErr)] = "Transport Header Format Err"
	gm[int(RmtHeaderFormatErr)] = "Header Format Err"
	gm[int(RmtTransportHeaderCRCErr)] = "Transport Header CRC Err"
	gm[int(RmtProtocolTypeErr)] = "Protocol Type Err"
	gm[int(RmtDiscardedSegment)] = "Discarded Segment"
	gm[int(RmtUnknownPortIdentifier)] = "Unknown Port"
}

type RadarMetrics struct {
	RadarIP utils.IP4
	Metrics
}

func (m *RadarMetrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"RadarIP": m.RadarIP,
		"Metrics": &m.Metrics,
	})
}

type globalRadarMetrics struct {
	Radar []RadarMetrics
}

func (g *globalRadarMetrics) ByIndex(index int) *RadarMetrics {
	return &g.Radar[index]
}

func (g *globalRadarMetrics) Init(radars int) {
	g.Radar = make([]RadarMetrics, radars)
	for i := range g.Radar {
		radar := &g.Radar[i]
		radar.Metrics.SetLength(int(radarMetricHead), int(radarMetricTail))
	}
}

func (g *globalRadarMetrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"RadarMetrics": g.Radar,
	})
}

var GlobalRadarMetrics globalRadarMetrics
