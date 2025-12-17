package service

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type SendQueueItem struct {
	RetryOn    time.Time
	CreateOn   time.Time
	MaxRetries int
	RetryNo    int
	Request    *port.Instruction
	Response   *port.Instruction
}

func (i *SendQueueItem) ShouldRetry() bool {
	return i.MaxRetries == 0 || i.RetryNo < i.MaxRetries
}

type instructionServiceStatus int

const (
	issAwaitSending instructionServiceStatus = iota
	issAwaitReceiving
)

// Instruction is meant to be specific to a radar
// Notes:
// 1. Only pop a message from the sendQueue if received
// 2. Always pop a message from receive queue
type Instruction struct {
	Context           any
	RadarIP           utils.IP4
	DataService       *UDPData
	sequenceNo        atomic.Uint32
	sendMutex         sync.Mutex
	sendQueue         utils.Queue
	receiveQueue      utils.Queue
	receiveMutex      sync.Mutex
	terminate         bool
	terminated        bool
	status            instructionServiceStatus
	OnIdle            func(*Instruction) bool
	OnResponse        func(*Instruction, *SendQueueItem)
	OnSequenceError   func(service *Instruction, instruction *port.Instruction, awaiting *SendQueueItem)
	OnAfterSendToUDP  func(*Instruction, utils.IP4, *port.Instruction)
	OnResend          func(*Instruction, *SendQueueItem)
	OnDropInstruction func(*Instruction, *SendQueueItem) bool
	IdleCooldownMs    time.Duration
	ResendsCooldownMs time.Duration
}

func (s *Instruction) Init() {
	s.ResendsCooldownMs = 60000
	s.IdleCooldownMs = 1000
}

// Start the Instruction Service
// dataService is used for writing the UDP to the radar
func (s *Instruction) Start(
	dataService *UDPData,
	radarIP utils.IP4,
) {
	if s.terminate {
		return
	}
	s.sendQueue.Init()
	s.receiveQueue.Init()
	s.DataService = dataService
	s.RadarIP = radarIP
	s.terminate = false
	s.terminated = false
	s.status = issAwaitSending
	s.sequenceNo.Store(0)
	go s.execute()
}

func (s *Instruction) EnqueueSend(instruction *port.Instruction, maxTries int) {
	now := time.Now()

	sqi := &SendQueueItem{
		MaxRetries: maxTries,
		RetryNo:    0,
		RetryOn:    now,
		CreateOn:   now,
		Request:    instruction,
		Response:   nil,
	}

	s.sendMutex.Lock()
	s.sendQueue.Push(sqi)
	s.sendMutex.Unlock()
}

func (s *Instruction) EnqueueReceive(instruction *port.Instruction) {
	s.receiveMutex.Lock()
	s.receiveQueue.Push(instruction)
	s.receiveMutex.Unlock()
}

func (s *Instruction) Stop() {
	s.terminate = true

	counter := 0
	for !s.terminated {
		counter++
		if counter > 30 { // willing to wait 3 seconds
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Instruction) execute() {
	for !s.terminate {
		var shouldSleep bool
		if s.status == issAwaitSending {
			shouldSleep = s.processSend()
		} else {
			shouldSleep = s.processReceive()
		}

		if shouldSleep {
			sleep := true

			if s.OnIdle != nil {
				sleep = s.OnIdle(s)
			}

			if sleep {
				time.Sleep(s.IdleCooldownMs * time.Millisecond)
			}
		}
	}
	s.terminated = true
}

func (s *Instruction) processSend() bool {
	s.sendMutex.Lock()
	front, ok := s.sendQueue.Peek()
	defer s.sendMutex.Unlock()

	if ok {
		iqi := front.(*SendQueueItem)

		if iqi.RetryNo == 0 {
			iqi.Request.Header.SequenceNo = s.nextSequenceNo()
		}
		iqi.RetryNo = iqi.RetryNo + 1
		iqi.RetryOn = time.Now()
		s.status = issAwaitReceiving

		// DataService should always be supplied
		slice := iqi.Request.SaveAsBytes()

		if s.DataService != nil {
			s.DataService.WriteData(s.RadarIP, slice)
		}

		if s.OnAfterSendToUDP != nil {
			s.OnAfterSendToUDP(s, s.RadarIP, iqi.Request)
		}
		return false
	}
	return true
}

func (s *Instruction) nextSequenceNo() uint32 {
	return s.sequenceNo.Add(1)
}

func (s *Instruction) processReceive() bool {
	var front interface{}
	var ok bool
	var ins *port.Instruction

	s.receiveMutex.Lock()
	defer s.receiveMutex.Unlock()

	if front, ok = s.receiveQueue.Pop(); ok {
		if ins, ok = front.(*port.Instruction); ok {
			return s.processReceiveInstruction(ins)
		}
	} else {
		// No message received, and we expect something
		// so lets check if you should resend the last sent message
		return s.processResend()
	}
	return true
}

func (s *Instruction) processResend() bool {
	s.sendMutex.Lock()
	send, ok := s.sendQueue.Peek()
	defer s.sendMutex.Unlock()

	if !ok {
		// Should never get here, this is simply as safety catch
		log.Warn().Msgf("process resend for %s is empty", s.RadarIP.String())
		return true
	}

	sentItem, ok := send.(*SendQueueItem)
	if !ok {
		// This can never happen, merely here as a precaution
		log.Error().Msgf("instruction service with invalid queue item type")
		return true
	}

	if utils.Time.IsOlderThan(sentItem.RetryOn, s.ResendsCooldownMs*time.Millisecond) {
		s.status = issAwaitSending
		if sentItem.ShouldRetry() {
			// By simple setting the status, the message will be resent
			if s.OnResend != nil {
				s.OnResend(s, sentItem)
			}
		} else {
			drop := true
			if s.OnDropInstruction != nil {
				drop = s.OnDropInstruction(s, sentItem)
			}
			if drop {
				s.sendQueue.Pop()
			}
		}
	}
	return false
}

func (s *Instruction) processReceiveInstruction(insReceived *port.Instruction) bool {
	s.sendMutex.Lock()
	send, ok := s.sendQueue.Peek()

	if !ok {
		s.sendMutex.Unlock()
		s.handleSequenceIssue(insReceived, nil)
		return true
	}

	sentItem, ok := send.(*SendQueueItem)
	if !ok {
		// This can never happen, merely here as a precaution
		s.sendMutex.Unlock()
		log.Error().Msgf("instruction service with invalid queue item type")
		return true
	}

	if sentItem.Request.Header.SequenceNo != insReceived.Header.SequenceNo {
		s.sendMutex.Unlock()
		s.handleSequenceIssue(insReceived, sentItem)
		return true
	}

	// At this point we have a valid response to a request
	sentItem.Response = insReceived
	s.sendQueue.Pop()
	s.sendMutex.Unlock()

	if s.OnResponse != nil {
		s.OnResponse(s, sentItem)
		s.status = issAwaitSending
	}
	return false
}

func (s *Instruction) handleSequenceIssue(insReceived *port.Instruction, sentItem *SendQueueItem) {
	if sentItem != nil {
		if s.OnSequenceError != nil {
			s.OnSequenceError(s, insReceived, sentItem)
		} else {
			log.Warn().Msgf(
				"Received instruction %d not in expected sequence %d",
				insReceived.Header.SequenceNo,
				sentItem.Request.Header.SequenceNo,
			)
		}
	}
}

func (s *Instruction) shouldSleep() bool {
	switch s.status {
	case issAwaitSending:
		return s.sendQueue.IsEmpty()

	case issAwaitReceiving:
		return s.receiveQueue.IsEmpty()

	default:
		return true
	}
}
