package instrumentation

type SDLCMetricType int

const (
	sdlcMetricHead SDLCMetricType = iota + MetricStartForSDLC
	SDLCWebAPIRequest
	SDLCWebAPIFailure
	SDLCWebAPISuccess
	SDLCUARTReads
	SDLCUARTReadBytes
	SDLCUARTWriteSuccess
	SDLCUARTWriteSuccessBytes
	SDLCUARTWriteEnqueued
	SDLCUARTWriteEnqueuedBytes
	SDLCUARTWriteDequeued
	SDLCUARTWriteDequeuedBytes
	SDLCUARTWriteSkips     // Skips writes due to write channel being full
	SDLCUARTWriteSkipBytes // Skips writes due to write channel being full
	SDLCUARTWriteErrors
	SDLCUARTWriteErrorBytes
	SDLCUARTReadError
	SDLCUARTStaticStatusRequest
	SDLCUARTStaticStatusResponse
	SDLCUARTBIUDiagnosticsRequest
	SDLCUARTBIUDiagnosticsResponse
	SDLCUARTSIUDiagnosticsRequest
	SDLCUARTSIUDiagnosticsResponse
	SDLCUARTDynamicStatusRequest
	SDLCUARTDynamicStatusResponse
	SDLCUARTDiagnosticsRequest
	SDLCUARTDiagnosticsResponse
	SDLCUARTSendDetectRequest
	SDLCUARTAcknowledgeResponse
	SDLCUARTCMUFrames
	SDLCUARTDateTimeFrames

	sdlcMetricTail
)

func init() {
	gm := GlobalMetricNames()
	gm[int(SDLCWebAPIRequest)] = "SDLCWebAPIRequest"
	gm[int(SDLCWebAPIFailure)] = "SDLCWebAPIFailure"
	gm[int(SDLCWebAPISuccess)] = "SDLCWebAPISuccess"
	gm[int(SDLCUARTWriteErrors)] = "SDLCUARTWriteErrors"
	gm[int(SDLCUARTWriteErrorBytes)] = "SDLCUARTWriteErrors"
	gm[int(SDLCUARTReadError)] = "SDLCUARTReadError"
	gm[int(SDLCUARTReads)] = "SDLCUARTReads"
	gm[int(SDLCUARTReadBytes)] = "SDLCUARTReadBytes"
	gm[int(SDLCUARTWriteSuccess)] = "SDLCUARTWriteSuccess"
	gm[int(SDLCUARTWriteEnqueued)] = "Writes Enqueued"
	gm[int(SDLCUARTWriteEnqueuedBytes)] = "Write Enqueued Bytes"
	gm[int(SDLCUARTWriteDequeued)] = "Writes Dequeued"
	gm[int(SDLCUARTWriteDequeuedBytes)] = "Write Dequeued Bytes"
	gm[int(SDLCUARTWriteSuccessBytes)] = "SDLCUARTWriteSuccessBytes"
	gm[int(SDLCUARTWriteSkips)] = "SDLCUARTWriteSkips"
	gm[int(SDLCUARTWriteSkipBytes)] = "SDLCUARTWriteSkipBytes"
	gm[int(SDLCUARTStaticStatusRequest)] = "SDLCUARTStaticStatusRequest"
	gm[int(SDLCUARTStaticStatusResponse)] = "SDLCUARTStaticStatusResponse"
	gm[int(SDLCUARTBIUDiagnosticsRequest)] = "SDLCUARTBIUDiagnosticsRequest"
	gm[int(SDLCUARTBIUDiagnosticsResponse)] = "SDLCUARTBIUDiagnosticsResponse"
	gm[int(SDLCUARTSIUDiagnosticsRequest)] = "SDLCUARTSIUDiagnosticsRequest"
	gm[int(SDLCUARTSIUDiagnosticsResponse)] = "SDLCUARTSIUDiagnosticsResponse"
	gm[int(SDLCUARTDynamicStatusRequest)] = "SDLCUARTDynamicStatusRequest"
	gm[int(SDLCUARTDynamicStatusResponse)] = "SDLCUARTDynamicStatusResponse"
	gm[int(SDLCUARTDiagnosticsRequest)] = "SDLCUARTDiagnosticsRequest"
	gm[int(SDLCUARTDiagnosticsResponse)] = "SDLCUARTDiagnosticsResponse"
	gm[int(SDLCUARTSendDetectRequest)] = "SDLCUARTSendDetectRequest"
	gm[int(SDLCUARTAcknowledgeResponse)] = "SDLCUARTAcknowledgeResponse"
	gm[int(SDLCUARTCMUFrames)] = "CMU"
	gm[int(SDLCUARTDateTimeFrames)] = "Date Time"

	GlobalSDLCMetrics.Metrics.SetLength(int(sdlcMetricHead), int(sdlcMetricTail))
}

type globalSDLCMetrics struct {
	Metrics Metrics
}

var GlobalSDLCMetrics globalSDLCMetrics

func (s *globalSDLCMetrics) Get(metric SDLCMetricType) *Metric {
	return s.Metrics.GetRel(int(metric))
}
