package integration

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"rvpro3/radarvision.com/internal/api/metrics"
	"rvpro3/radarvision.com/internal/api/radar"
	"rvpro3/radarvision.com/utils"
)

func TestHello(t *testing.T) {
	utils.Print.Ln("integration.TestHello")
}

func getAPIUrl() string {
	return utils.Args.GetString("-api", "http://127.0.0.1:8080")
}

func getRVMUrl() string {
	return utils.Args.GetString("-rvmIP", "127.0.0.1:55555")
}

func getRadarIP(radarNo int) string {
	name := fmt.Sprintf("-radar%dIP", radarNo)
	return utils.Args.GetString(name, "127.0.0.1:5000"+strconv.Itoa(radarNo))
}

// --args -radar1=127.0.0.1:50001 -api=http://127.0.0.1:8080
func TestTransportHeaderFormatErr(t *testing.T) {
	apiUrlStr := getAPIUrl()
	rvmIPStr := getRVMUrl()
	radarIPStr := getRadarIP(1)
	radarIP := utils.IP4Builder.FromString(radarIPStr)

	var err error
	var bef, aft *metrics.UDPBroker
	hello := "Hello World!"

	api := metrics.NewMetricsAPI(apiUrlStr)

	bef, err = api.GetUDPBroker(radarIP)
	utils.Debug.Panic(err)

	radarSim := radar.RadarSimulator{
		RadarIP4:  radarIP,
		ServerIP4: utils.IP4Builder.FromString(rvmIPStr),
	}

	utils.Debug.Panic(radarSim.SendStr(hello))

	aft, err = api.GetUDPBroker(radarIP)
	utils.Debug.Panic(err)

	assertBrokerDiff(t, radarIP, bef, aft, "TransportHeaderFormatErr", 1)
	assertBrokerDiff(t, radarIP, bef, aft, "ReceivePackets", 1)
	assertBrokerDiff(t, radarIP, bef, aft, "ReceiveBytes", int64(len(hello)))
}

func assertBrokerDiff(
	t *testing.T,
	radarIP utils.IP4,
	bef *metrics.UDPBroker,
	aft *metrics.UDPBroker,
	metricName string,
	difference int64,
) {
	utils.Test.Fmt("Testing radar %s, metric [%s] for delta %d\n", radarIP, metricName, difference)
	t1 := bef.GetBroker(radarIP, metricName)
	t2 := aft.GetBroker(radarIP, metricName)
	utils.Debug.PanicIf(t1 == nil || t2 == nil, "No data")
	//goland:noinspection GoMaybeNil
	assert.Equal(t, difference, t2.Value-t1.Value)
}
