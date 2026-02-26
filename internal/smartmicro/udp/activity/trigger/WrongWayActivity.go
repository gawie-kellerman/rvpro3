package trigger

import (
	"bytes"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/bit"
)

type wrongWayStatus int

const (
	wwsPending wrongWayStatus = iota
	wwsWriteCache
	wwsWriteProgress
	wwsWriteTail
)

type WrongWayActivity struct {
	InCase               bool
	CaseStartTime        time.Time
	CaseStatus           wrongWayStatus
	CasePath             string
	CaseID               string
	ChannelSO            int
	ChannelMask          uint64
	ChannelMaskPrevious  uint64
	StreamURL            string
	MaxHeadPhotos        int
	MaxRecordingDuration time.Duration
	jpegService          *CaptureMJPegService
	jpegCache            CaptureMJPegCache
	Metrics              WrongWayActivityMetrics `json:"-"`
	interfaces.UDPActivityMixin
}

type WrongWayActivityMetrics struct {
	PortVersionError           *utils.Metric
	StatePendingWithoutTrigger *utils.Metric
	utils.MetricsInitMixin
}

func (w *WrongWayActivity) Init(workflow interfaces.IUDPWorkflow, index int, metricsName string) {
	w.InitBase(workflow, index, metricsName)
	w.Metrics.InitMetrics(metricsName, &w.Metrics)

	w.jpegService = &CaptureMJPegService{
		StreamURL:       w.StreamURL,
		Enabled:         true,
		OnFrameCallback: w.onFrameCallback,
		OnErrorCallback: w.onErrorCallback,
	}
	w.jpegCache.Init(w.MaxHeadPhotos)
	w.jpegService.SetupAndStart(&utils.GlobalState, &utils.GlobalSettings)
	w.ChannelMask = bit.Set(uint64(0), w.ChannelSO)
}

func (w *WrongWayActivity) Process(now time.Time, buffer []byte) {
	th := port.TransportHeaderReader{Buffer: buffer}
	ph := port.PortHeaderReader{Buffer: buffer, StartOffset: int(th.GetHeaderLength())}

	if ph.GetPortMajorVersion() == 2 && ph.GetPortMinorVersion() == 0 {
		trg := port.EventTriggerReader{}
		trg.Init(buffer[:])

		relays := trg.GetRelays()
		currentMask := relays & w.ChannelMask

		switch w.CaseStatus {
		case wwsPending:
			w.handlePending(now, trg, currentMask)

		case wwsWriteCache:
			//w.handleWriteCache(trg, currentMask)

		case wwsWriteProgress:
			//w.handleProgress(trg, currentMask)

		case wwsWriteTail:
			//w.handleWriteTail(trg, currentMask)
		}

	} else {
		w.Metrics.PortVersionError.IncAt(1, now)
	}
}

func (w *WrongWayActivity) handlePending(now time.Time, trg port.EventTriggerReader, mask uint64) {
	if mask == 0 {
		w.Metrics.StatePendingWithoutTrigger.IncAt(1, time.Now())
		return
	}
	w.startTransaction(now, trg)
}

func (w *WrongWayActivity) startTransaction(now time.Time, trg port.EventTriggerReader) {
	w.CaseStatus = wwsWriteCache
}

func (w *WrongWayActivity) onFrameCallback(service any, now time.Time, buffer *bytes.Buffer) {

}

func (w *WrongWayActivity) onErrorCallback(service any, now time.Time, err error) {
}
