package uartsdlc

import (
	"sync"
)

type SDLCWritePool struct {
	pool sync.Pool
}

func NewSDLCWritePool() *SDLCWritePool {
	return &SDLCWritePool{
		pool: sync.Pool{
			New: func() any {
				return make([]byte, 128)
			},
		},
	}
}

func (p *SDLCWritePool) Alloc() []byte {
	return p.pool.Get().([]byte)
}

func (p *SDLCWritePool) Release(buffer []byte) {
	p.pool.Put(buffer)
}
