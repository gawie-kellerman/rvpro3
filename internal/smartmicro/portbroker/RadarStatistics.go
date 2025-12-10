package portbroker

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/common"
	"rvpro3/radarvision.com/utils"
)

type RsErrorType int

const (
	RsHead RsErrorType = iota + 200
	RsTotalMessages
	RsObjectListCount
	RsStatisticsCount
	RsTriggerCount
	RsPVRCount
	RsInstructionCount
	RsDiagnosticCount
	RsTransportHeaderFormatErr
	RsPortHeaderFormatErr
	RsTransportHeaderCRCErr
	RsPortHeaderCRCErr
	RsProtocolTypeErr
	RsDiscardedSegment
	RsUnknownPortIdentifier
	RsMessageDrop
	RsSegmentationBufferOverflow
	RsTail
)

type RadarStatistics struct {
	RadarIP          utils.IP4
	Data             [RsTail - RsHead - 1]common.CounterStatistic
	ActiveStatsCount int
}

func (stats *RadarStatistics) Init(radarIP utils.IP4) {
	stats.RadarIP = radarIP
	for i := 0; i < len(stats.Data); i++ {
		stat := &stats.Data[i]
		stat.Id = int(RsHead) + i + 1
		stat.IsActive = true
	}
}

func (stats *RadarStatistics) WriteToFixedBuffer(writer *utils.FixedBuffer) {
	if stats.ActiveStatsCount == 0 {
		stats.ActiveStatsCount = common.StatsHelper.CountActives(stats.Data[:])
	}

	stats.RadarIP.WriteToFixedBuffer(writer)
	writer.WriteU8(uint8(stats.ActiveStatsCount))

	start := int(RsHead) + 1
	for i := start; i < int(RsTail); i++ {
		stat := &stats.Data[i-start]
		stat.WriteToFixedBuffer(writer)
	}
}

// Register returns a true if it is a first time registry
func (stats *RadarStatistics) Register(errNo RsErrorType, now time.Time) bool {
	data := &stats.Data[errNo-RsHead]
	return data.Add(1, now)
}
