package broker

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"
)

func TestMessageHub_Handle(t *testing.T) {
	var buffer [200]byte
	hub := MessageHub{}

	hub.Init()

	res := testing.Benchmark(func(b *testing.B) {
		for n := 0; n < 1000; n++ {
			now := time.Now()
			binary.LittleEndian.PutUint64(buffer[0:8], uint64(now.Unix()))
			hub.Handle(n%4, now, 1, buffer[:8])
			time.Sleep(time.Millisecond * 10)
		}
	})
	time.Sleep(3 * time.Second)
	fmt.Println(hub.Handlers[0].Test.Count)
	fmt.Println(res.MemAllocs)
	fmt.Println(res.Bytes)
	fmt.Println(res.AllocsPerOp())
	fmt.Println(res.AllocedBytesPerOp())
	fmt.Println(res.MemBytes)
}
