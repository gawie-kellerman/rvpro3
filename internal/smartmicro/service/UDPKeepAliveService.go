package service

import (
	"net"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

// UDPKeepAliveService to keep the radar alive
type UDPKeepAliveService struct {
	MixinDataService
	ClientId        uint32
	SourceIPAddr    string
	TargetIPAddr    string
	MulticastIPAddr string
	CooldownMs      int
	TimeoutMs       int
	IsActive        bool
	connection      *net.UDPConn
	buffer          [34]byte
	bufferLen       int
}

func NewT44KeepAliveService() *UDPKeepAliveService {
	return &UDPKeepAliveService{
		CooldownMs: 1000,
		TimeoutMs:  1000,
		IsActive:   true,
		MixinDataService: MixinDataService{
			LoopGuard:  utils.InfiniteLoopGuard{},
			RetryGuard: utils.RetryGuard{},
		},
	}
}

func (s *UDPKeepAliveService) Init() {
	s.CooldownMs = 1000
	s.TimeoutMs = 1000
	s.IsActive = true
	s.MulticastIPAddr = "239.144.0.0:60000"
	s.LoopGuard = &utils.InfiniteLoopGuard{}
	s.RetryGuard = utils.RetryGuard{
		ModCycles: 3,
	}
}

func (s *UDPKeepAliveService) Execute() {
	s.initBuffer()
	s.OnStartCallback(s)

	for s.Terminating = false; !s.Terminating; {
		s.now = time.Now()

		if s.IsActive {
			if s.openConnection() {
				s.sendAlive()
			}
		}
		s.Terminating = !s.LoopGuard.ShouldContinue(s.now)

		if !s.Terminating {
			s.onLoopCallback(s)
			time.Sleep(time.Duration(s.CooldownMs) * time.Millisecond)
		}
	}
	s.closeConnection()
	s.Terminated = true
	s.OnTerminateCallback(s)
}

func (s *UDPKeepAliveService) initBuffer() {
	targetIP, err := net.ResolveUDPAddr("udp4", s.TargetIPAddr)

	if err != nil {
		s.OnErrorCallback(s, err)
		return
	}

	ip4 := utils.IP4Builder.FromString(targetIP.String())

	alive := port.NewClientKeepAlive(
		s.ClientId,
		ip4.ToU32(),
		uint16(targetIP.Port),
	)

	writer := utils.NewFixedBuffer(s.buffer[:], 0, 0)
	alive.Write(&writer)
	s.bufferLen = writer.WritePos
}

func (s *UDPKeepAliveService) openConnection() bool {
	var err error
	var multicastAddr *net.UDPAddr
	var targetAddr *net.UDPAddr
	var ip4 utils.IP4

	if s.connection != nil {
		return true
	}

	if !s.RetryGuard.ShouldRetry() {
		return false
	}

	if multicastAddr, err = net.ResolveUDPAddr("udp4", s.MulticastIPAddr); err != nil {
		goto errorLabel
	}

	ip4 = utils.IP4Builder.
		FromString(s.TargetIPAddr).
		WithPort(0)

	if targetAddr, err = net.ResolveUDPAddr("udp4", ip4.ToString()); err != nil {
		goto errorLabel
	}

	if s.connection, err = net.DialUDP("udp4", targetAddr, multicastAddr); err != nil {
		goto errorLabel
	}

	s.RetryGuard.Reset()
	s.OnConnectCallback(s)

	return true

errorLabel:
	s.OnErrorCallback(s, err)
	s.closeConnection()
	return false
}

func (s *UDPKeepAliveService) sendAlive() {
	var err error
	timeout := s.now.Add(time.Duration(s.TimeoutMs) * time.Millisecond)
	err = s.connection.SetWriteDeadline(timeout)
	if err != nil {
		goto errLabel
	}
	if _, err = s.connection.Write(s.buffer[:s.bufferLen]); err != nil {
		goto errLabel
	}
	return

errLabel:
	s.OnErrorCallback(s, err)
	s.closeConnection()
	return
}

func (s *UDPKeepAliveService) closeConnection() {
	if s.connection != nil {
		s.OnDisconnectCallback(s)
		_ = s.connection.Close()
		s.connection = nil
	}
}
