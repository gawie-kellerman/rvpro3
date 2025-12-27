package brokerdeprecated

import "sync"

type MessagePool struct {
	pool sync.Pool
}

func NewMessagePool() *MessagePool {
	return &MessagePool{
		pool: sync.Pool{
			New: func() any {
				return new(Message)
			},
		},
	}
}

func (m *MessagePool) Alloc() *Message {
	return m.pool.Get().(*Message)
}

func (m *MessagePool) Release(msg *Message) {
	m.pool.Put(msg)
}
