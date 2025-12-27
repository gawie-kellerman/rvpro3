package brokerdeprecated

import (
	"encoding/binary"
	"fmt"
	"time"
)

type TestHandler struct {
	Count uint64
}

func (handler *TestHandler) Handle(msg *Message) {
	now := uint64(time.Now().Unix())
	then := binary.LittleEndian.Uint64(msg.Data[:])

	handler.Count++
	if now+10 < then {
		fmt.Println("then :", then, " now :", now)
	}
}
