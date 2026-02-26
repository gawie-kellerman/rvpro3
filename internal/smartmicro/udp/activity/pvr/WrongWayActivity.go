package pvr

import (
	"bytes"
	"time"

	"github.com/google/uuid"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type wrongWayStatus int

const (
	wwsPending wrongWayStatus = iota
	wwsTakePhotoCache
	wwsRunning
)

type trackedObject struct {
	Id      uint8
	Zone    uint8
	Heading float32
	Class   port.ObjectClassType
	Length  float32
	Speed   float32
	StartAt time.Time
	StopAt  time.Time
}

type WrongWayActivity struct {
	CaseStartTime    time.Time
	CasePath         time.Time
	CaseID           string
	Status           wrongWayStatus
	StreamURL        string
	HeadLen          int
	TailLen          int
	jpegService      *CaptureMJPegService
	jpegCache        CaptureMJPegCache
	TailCount        int
	TrackedObjects   []trackedObject
	Metrics          WrongWayActivityMetrics `json:"-"`
	TrackedZones     []uint8
	MaxTimeInRunning time.Duration
	interfaces.UDPActivityMixin
}

type WrongWayActivityMetrics struct {
	CacheCount            *utils.Metric
	CaptureCount          *utils.Metric
	ClearCount            *utils.Metric
	PortVersionErrorCount *utils.Metric
	PendingNoObjects      *utils.Metric
	PendingWithObjects    *utils.Metric
	RunningNoObjects      *utils.Metric
	RunningWithObjects    *utils.Metric
	utils.MetricsInitMixin
}

// Init does not initialize StreamURL, HeadLen, or TailLen...
// These must be separately AND BEFORE Init is called
func (w *WrongWayActivity) Init(workflow interfaces.IUDPWorkflow, index int, metricsName string) {
	w.InitBase(workflow, index, metricsName)
	w.Metrics.InitMetrics(metricsName, &w.Metrics)

	w.TrackedObjects = make([]trackedObject, 0, 256)
	w.jpegService = &CaptureMJPegService{
		StreamURL:       w.StreamURL,
		Enabled:         true,
		OnFrameCallback: w.onFrameCallback,
		OnErrorCallback: w.onErrorCallback,
	}

	w.jpegCache.Init(w.HeadLen + w.TailLen)
	w.jpegService.SetupAndStart(&utils.GlobalState, &utils.GlobalSettings)
}

func (w *WrongWayActivity) Process(workflow interfaces.IUDPWorkflow, index int, now time.Time, bytes []byte) {
	th := port.TransportHeaderReader{Buffer: bytes}
	ph := port.PortHeaderReader{Buffer: bytes, StartOffset: int(th.GetHeaderLength())}

	if ph.GetPortMinorVersion() == 2 && ph.GetPortMinorVersion() == 0 {
		pvr := port.PVRReader{}
		pvr.Init(bytes[:])

		switch w.Status {
		case wwsPending:
			w.startTransaction(now, pvr)

		case wwsTakePhotoCache:
			w.monitorPhotoCache(now, pvr)

		case wwsRunning:
			w.continueTransaction(now, pvr)
		}
	} else {
		w.Metrics.PortVersionErrorCount.IncAt(1, now)
	}
}

func (w *WrongWayActivity) onFrameCallback(service any, now time.Time, buffer *bytes.Buffer) {
	switch w.Status {
	case wwsPending:
		// Simply cache the photo
		w.jpegCache.Push(buffer)
		w.Metrics.CacheCount.SetAt(int64(w.jpegCache.Depth()), now)
		break

	case wwsTakePhotoCache:
		for w.jpegCache.Depth() > 0 {
			first := w.jpegCache.GetFront()
			if first != nil {
				w.saveOffenseImage(first.Time, first.Buffer, false)
			}
			w.jpegCache.PopFront()
		}

	case wwsRunning:
		w.saveOffenseImage(now, buffer, true)
	}

	utils.Print.Ln("Photo ")
}

func (w *WrongWayActivity) onErrorCallback(service any, now time.Time, err error) {
	w.clearTransaction(now)
}

func (w *WrongWayActivity) clearTransaction(now time.Time) {
	w.TrackedObjects = w.TrackedObjects[:0]
	w.Metrics.ClearCount.IncAt(1, now)
	w.jpegCache.Clear()
	w.Status = wwsRunning
}

// captureTransaction must:
// 1. Create an offense by serial
// 2. Persist the offense (SQLite and storage)
// 3. Optionally send to down line (e.g. Sunguide)
func (w *WrongWayActivity) captureTransaction(now time.Time) {
	w.Metrics.CaptureCount.IncAt(1, now)
}

// startTransaction
// 1. Create an offense by serial
// 2. Persist the offense (SQLite and storage)
// 3. Optionally send to down line (e.g. Sunguide)

func (w *WrongWayActivity) startTransaction(now time.Time, pvr port.PVRReader) {
	count := w.fillTracking(now, pvr)

	if count == 0 {
		// None of the objects in the interested zones, so don't start running
		return
	}

	// At this point, at least 1 applicable pvr object
	w.CaseStartTime = now
	w.Status = wwsTakePhotoCache

	w.CaseID = w.getCaseID(now)
}

func (w *WrongWayActivity) getCaseID(now time.Time) string {
	uid, err := uuid.NewRandom()

	if err != nil {
		return w.getAlternateCaseID(now)
	}

	return uid.String()
}

func (w *WrongWayActivity) getAlternateCaseID(now time.Time) string {
	return now.Format(utils.FileDateTimeMS)
}

func (w *WrongWayActivity) shouldStartTransaction(now time.Time, pvr port.PVRReader) bool {
	nofObjects := int(pvr.GetNofObjects())

	for n := 0; n < nofObjects; n++ {
		zoneId := pvr.GetZone(n)
		if w.isTrackedZone(zoneId) {
			return true
		}
	}

	return false
}

func (w *WrongWayActivity) fillTracking(now time.Time, pvr port.PVRReader) int {
	nofObjects := int(pvr.GetNofObjects())

	for n := 0; n < nofObjects; n++ {
		srcObjId := pvr.GetObjectId(n)
		zoneId := pvr.GetZone(n)
		if w.isTrackedZone(zoneId) {
			tgtObj := w.getTrackedObject(srcObjId)

			if tgtObj == nil {
				w.TrackedObjects = append(w.TrackedObjects, trackedObject{
					Id:      srcObjId,
					Zone:    pvr.GetZone(n),
					Heading: pvr.GetHeading(n),
					Class:   pvr.GetObjectClass(n),
					Length:  pvr.GetLength(n),
					Speed:   pvr.GetSpeed(n),
					StartAt: now,
				})
			} else {
				tgtObj.StopAt = now
			}
		}
	}
	return len(w.TrackedObjects)
}

func (w *WrongWayActivity) getTrackedObject(objId uint8) *trackedObject {
	for index := range w.TrackedObjects {
		trObj := &w.TrackedObjects[index]

		if trObj.Id == objId {
			return trObj
		}
	}
	return nil
}

func (w *WrongWayActivity) continueTransaction(now time.Time, pvr port.PVRReader) {
}

func (w *WrongWayActivity) countTrackedZone(pvr port.PVRReader) (res int) {
	for index := range w.TrackedObjects {
		obj := &w.TrackedObjects[index]

		for _, trackedZone := range w.TrackedZones {
			if trackedZone == obj.Zone {
				res++
			}
		}
	}
	return res
}

func (w *WrongWayActivity) isTrackedZone(zone uint8) bool {
	for index := range w.TrackedObjects {
		if w.TrackedObjects[index].Zone == zone {
			return true
		}
	}
	return false
}

func (w *WrongWayActivity) monitorPhotoCache(now time.Time, pvr port.PVRReader) {
	dur := now.Sub(w.CaseStartTime)

	if dur < w.MaxTimeInRunning {
		return
	}

	// The vehicle or pvr is stuck...  simply going to wrap it up

}

func (w *WrongWayActivity) saveOffenseImage(t time.Time, buffer *bytes.Buffer, isNonCache bool) {

}
