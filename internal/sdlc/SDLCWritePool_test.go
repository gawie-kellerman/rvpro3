package sdlc

import (
	"sync"
	"testing"
)

func BenchmarkSDLCWritePool_Alloc(b *testing.B) {
	b.ReportAllocs()
	wp := NewSDLCWritePool()
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go allocAndFree(wp, &wg)
	}

	wg.Wait()
}

func allocAndFree(wp *SDLCWritePool, wg *sync.WaitGroup) {
	for i := 0; i < 10; i++ {
		buf := wp.Alloc()
		wp.Release(buf)
	}

	wg.Done()
}
