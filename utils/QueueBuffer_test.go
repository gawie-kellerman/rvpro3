package utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sample = "1234567890"
const alphabet = "abcdefghijklmnopqrstuvwxyz"

func TestQueueBuffer_PushSize(t *testing.T) {
	qb := NewQueueBuffer(500)
	panicOnPush(qb.PushData([]byte(sample), false))
	assert.Equal(t, 10, qb.Size())
	assert.Equal(t, 0, qb.front)
	assert.Equal(t, 10, qb.back)
	assert.True(t, bytes.EqualFold(qb.GetDataSlice(), []byte(sample)))
}

func TestQueueBuffer_PushRepeated(t *testing.T) {
	qb := NewQueueBuffer(500)
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample[:5]), false))
	assert.Equal(t, 25, qb.Size())
	assert.Equal(t, 0, qb.front)
	assert.Equal(t, 25, qb.back)
}

func TestQueueBuffer_PopSize(t *testing.T) {
	qb := NewQueueBuffer(500)

	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample[:5]), false))
	panicOnErr(qb.PopSize(20))

	assert.Equal(t, 5, qb.Size())
	assert.Equal(t, 20, qb.front)
	assert.Equal(t, 25, qb.back)
}

func TestQueueBuffer_PopEmpty(t *testing.T) {
	qb := NewQueueBuffer(500)

	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample[:5]), false))
	panicOnErr(qb.PopSize(20))
	panicOnErr(qb.PopSize(5))

	assert.Equal(t, 0, qb.Size())
	assert.Equal(t, 0, qb.front)
	assert.Equal(t, 0, qb.back)
}

func TestQueueBuffer_PopData(t *testing.T) {
	qb := NewQueueBuffer(500)

	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample[:5]), false))
	panicOnErr(qb.PopSize(20))

	var copyBuf [5]byte
	panicOnErr(qb.PopData(copyBuf[:]))

	assert.Equal(t, 0, qb.Size())
	assert.Equal(t, 0, qb.front)
	assert.Equal(t, 0, qb.back)
}

func TestQueueBuffer_PopWithOptimize(t *testing.T) {
	qb := NewQueueBuffer(50)

	// Sample is 10 in size
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnErr(qb.PopSize(20))
	// 40 Avail, 20 in front and 20 at back

	assert.Equal(t, 20, qb.GetFrontAvail())
	assert.Equal(t, 20, qb.GetBackAvail())

	// Pushing the alphabet means that the Buffer must be optimized
	panicOnPush(qb.PushData([]byte(alphabet), false))
	assert.True(t, bytes.EqualFold(qb.GetDataSlice(), []byte(sample+alphabet)))
}

func TestQueueBuffer_ErrMessageWontFit(t *testing.T) {
	qb := NewQueueBuffer(50)

	// Sample is 10 in size
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))
	panicOnPush(qb.PushData([]byte(sample), false))

	// Alphabet won't fit anymore
	assert.Equal(t, PoMessageCannotFit, qb.PushData([]byte(alphabet), false))
}

func TestQueueBuffer_ErrMessageTooLarge(t *testing.T) {
	qb := NewQueueBuffer(50)
	assert.Equal(t, PoMessageTooLarge, qb.PushData([]byte(alphabet+alphabet), false))
}

func panicOnPush(outcome PushOutcome) {
	if !outcome.IsSuccess() {
		panic(outcome)
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
