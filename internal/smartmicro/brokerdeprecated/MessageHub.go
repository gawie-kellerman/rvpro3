package brokerdeprecated

import (
	"fmt"
	"time"
)

type MessageHub struct {
	Handlers [4]MessageHandler
}

func (s *MessageHub) Init() {
	for i := 0; i < len(s.Handlers); i++ {
		h := &s.Handlers[i]
		h.Init()
		go h.execute()
	}
}

func (s *MessageHub) Handle(
	radarIndex int,
	messageTime time.Time,
	messageCode int,
	messageData []byte,
) {
	handler := &s.Handlers[radarIndex]
	qLen := handler.QueueLen()
	qCap := handler.QueueCap()

	if qLen >= qCap {
		s.logChannelQueueFull()
		return
	}

	handler.AddMessage(messageTime, messageCode, messageData)
}

func (s *MessageHub) logChannelQueueFull() {
	fmt.Println("logChannelQueueFull")
}
