package service

import (
	"time"

	"go.bug.st/serial"
)

const serialBufferSize = 1024 * 8

type SerialDataService struct {
	MixinDataService
	BaudRate  int
	PortName  string
	OnData    func(service *SerialDataService, data []byte) []byte
	OnNoData  func(service *SerialDataService)
	port      serial.Port
	isOpen    bool
	buffer    [serialBufferSize]byte
	bufferOff int
}

func (s *SerialDataService) Execute() {
	s.initBuffer()
	s.OnStartCallback(s)

	for s.Terminate = false; !s.Terminated; {
		s.now = time.Now()

		if s.openConnection() {
			s.receiveData()
		}

		s.Terminate = !s.LoopGuard.ShouldContinue(s.now)

		if !s.Terminate {
			s.onLoopCallback(s)
		}
	}

	s.closeConnection()
	s.Terminated = true
	s.OnTerminateCallback(s)
}

func (s *SerialDataService) openConnection() bool {
	var err error

	if s.isOpen {
		return true
	}

	if !s.RetryGuard.ShouldRetry() {
		return false
	}

	mode := &serial.Mode{
		BaudRate: s.BaudRate,
	}

	s.port, err = serial.Open(s.PortName, mode)
	if err != nil {
		goto errorLabel
	}

	return true

errorLabel:
	s.OnErrorCallback(s, err)
	s.closeConnection()
	return false
}

func (s *SerialDataService) closeConnection() {
	if s.isOpen {
		s.OnDisconnectCallback(s)
		_ = s.port.Close()
		s.isOpen = false
		s.bufferOff = 0
	}
}

func (s *SerialDataService) initBuffer() {
	s.bufferOff = 0
}

func (s *SerialDataService) receiveData() {
	var bytesRead int
	var err error

	bytesRead, err = s.port.Read(s.buffer[s.bufferOff:])

	if err != nil {
		s.OnErrorCallback(s, err)
		s.closeConnection()
	} else {
		s.bufferOff += bytesRead

		if s.OnData != nil {
			partial := s.OnData(s, s.buffer[:s.bufferOff])

			if len(partial) > 0 {
				copy(s.buffer[0:], partial)
				s.bufferOff = len(partial)
			} else {
				s.bufferOff = 0
			}
		}
	}
}
