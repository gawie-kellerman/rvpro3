package tcphub

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData [15]byte

func BenchmarkBuffer(t *testing.B) {
	var pk1Buffer [1024]byte
	copy(testData[:], "hello1234567890")

	buf := NewBuffer(300000)

	for n := 0; n < 1000; n++ {
		pk1 := NewPacket(testData[:rand.Intn(len(testData))])
		pk1Slice, _ := pk1.SaveToBytes(pk1Buffer[:])

		buf.PushBytes(pk1Slice)
		//buf.DumpSummary()
	}

	for n := 0; n < 1000; n++ {
		var pk2 Packet
		success := buf.Pop(&pk2)
		if !success {
			t.Fatal("unable to pop")
		}
	}

	assert.Zero(t, buf.Size())
}

func BenchmarkZeroBuffer(t *testing.B) {
	var pk1Buffer [1024]byte
	copy(testData[:], "hello1234567890")

	pk1 := NewPacket(testData[:rand.Intn(len(testData))])
	pk1Slice, _ := pk1.SaveToBytes(pk1Buffer[:])

	buf := NewBuffer(300000)

	for n := 0; n < 1000; n++ {

		buf.PushBytes(pk1Slice)

		//buf.DumpSummary()
		var pk2 Packet
		success := buf.Pop(&pk2)
		if !success {
			t.Fatal("pop error")
		}
	}

	assert.Zero(t, buf.Size())
}
