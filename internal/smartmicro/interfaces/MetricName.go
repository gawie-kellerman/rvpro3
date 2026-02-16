package interfaces

import (
	"fmt"

	"rvpro3/radarvision.com/utils"
)

func GetUDPRadarMetric(radarIP4 utils.IP4) string {
	return fmt.Sprintf("UDP.Broker.[%s]", radarIP4)
}
