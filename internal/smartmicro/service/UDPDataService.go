package service

import (
	"net"
	"time"

	"rvpro3/radarvision.com/utils"
)

const udpBufferSize = 8192

// UDPDataService has no IsActive for, if there is
// no KeepAlive then there is no data
type UDPDataService struct {
	MixinDataService
	ServerIPAddr string
	OnData       func(*UDPDataService, *net.UDPAddr, []byte)
	buffer       [udpBufferSize]byte
	bufferLen    int
	connection   *net.UDPConn
}

func (s *UDPDataService) Init() {
	s.LoopGuard = utils.InfiniteLoopGuard{}
	s.RetryGuard = utils.RetryGuard{
		ModCycles: 5,
	}
}

func (s *UDPDataService) Execute() {
	s.OnStartCallback(s)

	for s.Terminating = false; !s.Terminating; {
		s.now = time.Now()

		if s.openConnection() {
			s.receiveData()
		}

		s.Terminating = !s.LoopGuard.ShouldContinue(s.now)

		if !s.Terminating {
			s.onLoopCallback(s)
		}
	}

	s.closeConnection()
	s.Terminated = true
	s.OnTerminateCallback(s)
}

func (s *UDPDataService) openConnection() bool {
	var err error
	var serverAddr *net.UDPAddr

	if s.connection != nil {
		return true
	}

	if !s.RetryGuard.ShouldRetry() {
		return false
	}

	if serverAddr, err = net.ResolveUDPAddr("udp4", s.ServerIPAddr); err != nil {
		goto onErrorLabel
	}

	if s.connection, err = net.ListenUDP("udp4", serverAddr); err != nil {
		goto onErrorLabel
	}

	if err = s.connection.SetReadBuffer(udpBufferSize); err != nil {
		goto onErrorLabel
	}

	s.RetryGuard.Reset()
	s.OnConnectCallback(s)

	return true

onErrorLabel:
	s.OnErrorCallback(s, err)
	s.closeConnection()
	return false
}

func (s *UDPDataService) receiveData() {
	var err error
	var fromAddr *net.UDPAddr

	deadline := time.Now().Add(time.Duration(3) * time.Second)

	if err = s.connection.SetReadDeadline(deadline); err != nil {
		goto onErrorLabel
	}

	if s.bufferLen, fromAddr, err = s.connection.ReadFromUDP(s.buffer[:]); err != nil {
		goto onErrorLabel
	}

	if s.OnData != nil {
		s.OnData(s, fromAddr, s.buffer[:s.bufferLen])
	}

	return

onErrorLabel:
	s.OnErrorCallback(s, err)
	s.closeConnection()
}

func (s *UDPDataService) closeConnection() {
	if s.connection != nil {
		s.OnDisconnectCallback(s)
		_ = s.connection.Close()
		s.connection = nil
	}
}
