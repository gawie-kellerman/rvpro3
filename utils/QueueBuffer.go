package utils

import "errors"

type QueueBuffer struct {
	buffer []byte
	front  int
	back   int
}

var ErrBufferReadOverflow = errors.New("message read overflow")

type PushOutcome int

const (
	PoSuccess PushOutcome = iota
	PoBufferReset
	PoMessageTooLarge
	PoMessageCannotFit
)

func (p PushOutcome) IsSuccess() bool {
	return int(p) <= int(PoBufferReset)
}

func NewQueueBuffer(size int) *QueueBuffer {
	return &QueueBuffer{
		buffer: make([]byte, size),
		front:  0,
		back:   0,
	}
}

func (qb *QueueBuffer) Init(size int) {
	if cap(qb.buffer) != size {
		qb.buffer = make([]byte, size)
	}

	qb.front = 0
	qb.back = 0
}

func (qb *QueueBuffer) Capacity() int {
	return cap(qb.buffer)
}

func (qb *QueueBuffer) GetBackAvail() int {
	return cap(qb.buffer) - qb.back
}

func (qb *QueueBuffer) GetFrontAvail() int {
	return qb.front
}

func (qb *QueueBuffer) GetTotalAvail() int {
	return qb.GetFrontAvail() + qb.GetBackAvail()
}

func (qb *QueueBuffer) Size() int {
	return qb.back - qb.front
}

func (qb *QueueBuffer) Reset() {
	qb.front = 0
	qb.back = 0
}

func (qb *QueueBuffer) GetDataSlice() []byte {
	return qb.buffer[qb.front:qb.back]
}

func (qb *QueueBuffer) GetAvailSlice() []byte {
	return qb.buffer[qb.back:]
}

func (qb *QueueBuffer) PushSize(size int, force bool) PushOutcome {
	if outcome := qb.EnsureFit(size, force); !outcome.IsSuccess() {
		return outcome
	}
	qb.back += size
	return PoSuccess
}

// PushData
// Use force parameter to reset the Buffer if the Buffer is large enough
// to contain the data but does not have sufficient space left
// Returns
func (qb *QueueBuffer) PushData(data []byte, force bool) PushOutcome {
	if outcome := qb.EnsureFit(len(data), force); !outcome.IsSuccess() {
		return outcome
	} else {
		copy(qb.buffer[qb.back:qb.back+len(data)], data)
		qb.back += len(data)
		return outcome
	}
}

func (qb *QueueBuffer) PopSize(size int) error {
	if qb.Size() < size {
		return ErrBufferReadOverflow
	}

	qb.front += size
	qb.optimize()

	return nil
}

func (qb *QueueBuffer) PopData(target []byte) error {
	if qb.Size() < len(target) {
		return ErrBufferReadOverflow
	}

	qb.front += len(target)
	qb.optimize()

	return nil
}

func (qb *QueueBuffer) optimize() {
	if qb.front == qb.back {
		qb.front = 0
		qb.back = 0
	}
}

func (qb *QueueBuffer) GetBackSlice() []byte {
	return qb.buffer[qb.back:]
}

func (qb *QueueBuffer) Optimize() {
	size := qb.Size()

	if size > 0 && qb.front > 0 {
		copy(qb.buffer[0:], qb.buffer[qb.front:qb.back])
		qb.front = 0
		qb.back = size
	}
}

func (qb *QueueBuffer) EnsureFit(sizeNeeded int, force bool) PushOutcome {
	// The message can fit at the back
	if sizeNeeded <= qb.GetBackAvail() {
		return PoSuccess
	}

	// Message cannot ever fit
	if sizeNeeded > qb.Capacity() {
		return PoMessageTooLarge
	}

	// The message can fit if Buffer is optimized
	if sizeNeeded <= qb.GetTotalAvail() {
		qb.Optimize()
		return PoSuccess
	}

	// Force the message in by resetting the Buffer
	if force {
		qb.Reset()
		return PoBufferReset
	}

	return PoMessageCannotFit
}
