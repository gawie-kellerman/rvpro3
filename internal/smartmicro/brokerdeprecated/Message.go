package brokerdeprecated

import "time"

type Message struct {
	ReceiveOn   time.Time
	MessageCode int
	Data        [4096]byte
	DataLen     int
	CompleteOn  time.Time
}

func (m *Message) GetReceiveOn() time.Time {
	return m.ReceiveOn
}

func (m *Message) GetMessageCode() int {
	return m.MessageCode
}
