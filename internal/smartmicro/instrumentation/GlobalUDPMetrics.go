package instrumentation

type UDPMetric int

const (
	udpMetricHead UDPMetric = iota + MetricStartForUDP
	UDPNow
	UDPMetricIncorrectRadar
	UDPMetricSocketOpen
	// UDPMetricSocketOpenFail includes both failure when opening the socket or
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

	gm := GlobalMetricNames()
	gm[int(UDPNow)] = "now"
	gm[int(UDPMetricIncorrectRadar)] = "error: incorrect radar"
	gm[int(UDPMetricSocketOpen)] = "socket open"
	gm[int(UDPMetricSocketOpenFail)] = "error: socket open fail"
	gm[int(UDPMetricSocketReadFail)] = "error: socket read fail"
	gm[int(UDPMetricSocketWriteFail)] = "error: socket write fail"
	gm[int(UDPMetricSocketUse)] = "socket reuse"
	gm[int(UDPMetricSocketSkip)] = "error: skip failed socket"
	gm[int(UDPMetricDataIterations)] = "data iterations"
	gm[int(UDPMetricNoDataReceived)] = "no data received"
	gm[int(UDPMetricDataBytes)] = "data bytes"
}
