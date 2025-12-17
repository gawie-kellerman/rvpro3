package instrumentation

type UDPMetric int

const (
	udpMetricHead UDPMetric = iota + 100
	UDPNow
	UDPMetricIncorrectRadar
	UDPMetricSocketOpen
	// UDPMetricSocketFail includes both failure when opening the socket or
	// socket read or writes
	UDPMetricSocketOpenFail
	UDPMetricSocketReadFail
	UDPMetricSocketWriteFail
	UDPMetricSocketUse
	UDPMetricSocketSkip
	UDPMetricDataIterations
	UDPMetricNoDataReceived
	UDPMetricDataBytes
	udpMetricTail
)

var GlobalUDPMetrics Metrics

func init() {
	GlobalUDPMetrics.SetLength(int(udpMetricHead), int(udpMetricTail))
}
