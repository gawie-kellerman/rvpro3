package service

import (
	"fmt"
	"testing"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

const testQty = 1000

func TestInstructService(t *testing.T) {
	service := Instruction{}
	service.Init()
	service.Start(nil, utils.IP4Builder.FromString("192.168.11.12:55555"))

	successCount := 0
	sequenceCount := 0
	writeCount := 0
	extraCount := 0
	resendCount := 0
	dropCount := 0

	service.OnDropInstruction = func(service *Instruction, item *SendQueueItem) bool {
		dropCount++
		fmt.Println("dropped ", item.Request.Header.SequenceNo)
		return true
	}

	service.OnResend = func(service *Instruction, item *SendQueueItem) {
		resendCount++
		fmt.Println("resend : ", item.Request.Header.SequenceNo)
	}

	service.OnResponse = func(s *Instruction, i *SendQueueItem) {
		successCount++

		if successCount%250 == 0 {
			extraCount++
			s.EnqueueSend(port.NewInstruction(), i.MaxRetries)
		}
	}

	service.OnSequenceError = func(service *Instruction, instruction *port.Instruction, awaiting *SendQueueItem) {
		sequenceCount++
	}

	service.OnAfterSendToUDP = func(service *Instruction, ip4 utils.IP4, instruction *port.Instruction) {
		writeCount++
		fmt.Println("Write ", instruction.Header.SequenceNo)
	}

	go enqueueSends(&service)
	go enqueueReceives(&service)

	time.Sleep(120 * time.Second)
	fmt.Println("Receive Success count:", successCount)
	fmt.Println("Sequence issue count:", sequenceCount)
	fmt.Println("Resend Data:", resendCount)
	fmt.Println("Write instruction count:", writeCount)
	fmt.Println("Drop instruction count:", dropCount)
	fmt.Println("Extra instruction count:", extraCount)
	fmt.Println("Send Queue Depth:", service.sendQueue.Len())
	fmt.Println("Receive Queue Depth:", service.receiveQueue.Len())
}

func enqueueReceives(service *Instruction) {
	for n := 1; n <= testQty; n++ {
		ins := port.NewInstruction()
		ins.Header.SequenceNo = uint32(n)
		service.EnqueueReceive(ins)

		if n%10 == 0 {
			ins = port.NewInstruction()
			ins.Header.SequenceNo = uint32(n)
			service.EnqueueReceive(ins)
		}
	}
}

func enqueueSends(service *Instruction) {
	ins := port.NewInstruction()
	for n := 1; n <= testQty; n++ {
		service.EnqueueSend(ins, 3)
	}
}
