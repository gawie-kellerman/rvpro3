package broker

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
)

const MhAbort = 0
const MhTest = 1

//const

type MessageHandler struct {
	channel     chan *Message
	pool        *MessagePool
	Terminating bool
	Terminated  bool
	Stats       StatisticsHandler
	ObjList     ObjectListHandler
	Trigger     TriggerHandler
	Instruction InstructionHandler
	Test        TestHandler
}

func (h *MessageHandler) AddMessage(receiveOn time.Time, messageCode int, data []byte) {
	message := h.pool.Alloc()
	message.ReceiveOn = receiveOn
	message.MessageCode = messageCode
	message.DataLen = len(data)
	copy(message.Data[:], data)

	h.channel <- message
}

func (h *MessageHandler) Init() {
	h.channel = make(chan *Message, 5)
	h.pool = NewMessagePool()
}

func (h *MessageHandler) Execute() {

	for h.Terminating = false; !h.Terminating; {
		select {
		case msg := <-h.channel:
			h.handleMessage(msg)
		}
	}
	close(h.channel)

	h.Terminated = true
}

func (h *MessageHandler) handleMessage(msg *Message) {
	switch msg.GetMessageCode() {
	case MhAbort:
		h.Terminating = true

	case MhTest:
		h.Test.Handle(msg)

	case port.PiObjectList:
		h.ObjList.Handle(msg)

	case port.PiEventTrigger:
		h.Trigger.Handle(msg)

	case port.PiInstruction:
		h.Instruction.Handle(msg)

	case port.PiStatistics:
		h.Stats.Handle(msg)
	}

	h.pool.Release(msg)
}

func (h *MessageHandler) QueueLen() int {
	return len(h.channel)
}

func (h *MessageHandler) QueueCap() int {
	return cap(h.channel)
}
