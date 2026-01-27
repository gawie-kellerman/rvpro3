package interfaces

import (
	"fmt"

	"rvpro3/radarvision.com/utils"
)

const MessageProcessedOnMetricName = "Message Processed On"

type metricName struct {
}

var MetricName metricName

func (metricName) GetUDPRadarMetric(radarIP4 utils.IP4) string {
	return fmt.Sprintf("Radar.%s", radarIP4.ToIPString())
}
