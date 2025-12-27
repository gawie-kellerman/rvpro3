package service

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

// UDPKeepAlive to keep the radar alive
type UDPKeepAlive struct {
	ClientId         uint32
	LocalIPAddr      utils.IP4
	MulticastIPAddr  utils.IP4
	CooldownMs       int
	ReconnectOnCycle int
	SendTimeout      int
	connection       utils.UDPClientConnection
	buffer           [34]byte
	bufferLen        int
	terminate        bool
	terminated       bool
	now              time.Time
	OnTerminate      func(*UDPKeepAlive)
	utils.ErrorLoggerMixin
}

func (s *UDPKeepAlive) Init() {
	s.CooldownMs = 1000
	s.SendTimeout = 1000
	s.ClientId = 0x1000001
	s.MulticastIPAddr = utils.IP4Builder.FromString("239.144.0.0:60000")
	s.LocalIPAddr = utils.IP4Builder.FromString("192.168.11.2:55555")
	s.ReconnectOnCycle = 5
}

func (s *UDPKeepAlive) Start() {
	s.terminate = false
	s.terminated = false
	s.connection.OnError = s.onConnectionError
	s.connection.Init(s.LocalIPAddr.WithPort(0), s.MulticastIPAddr, s, s.ReconnectOnCycle)

	go s.executeWrite()
}

func (s *UDPKeepAlive) onConnectionError(connection *utils.UDPClientConnection, err error) {
	s.LogError("UDPKeepAlive", err)
}

func (s *UDPKeepAlive) Stop() {
	s.terminate = true
	for !s.terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *UDPKeepAlive) Run() {
	go s.executeWrite()
}

func (s *UDPKeepAlive) executeWrite() {
	s.initBuffer()

	for !s.terminate {
		s.now = time.Now()

		if s.connection.Connect() {
			s.sendAlive()
		}

		if !s.terminate {
			time.Sleep(time.Duration(s.CooldownMs) * time.Millisecond)
		}
	}
	s.connection.Disconnect()
	s.terminated = true
	if s.OnTerminate != nil {
		s.OnTerminate(s)
	}
}

func (s *UDPKeepAlive) initBuffer() {
	alive := port.NewClientKeepAlive(
		s.ClientId,
		s.LocalIPAddr.ToU32(),
		uint16(s.LocalIPAddr.Port),
	)

	writer := utils.NewFixedBuffer(s.buffer[:], 0, 0)
	alive.Write(&writer)
	s.bufferLen = writer.WritePos
}

func (s *UDPKeepAlive) sendAlive() {
	cnx := s.connection.GetConnection()

	if cnx == nil {
		return
	}

	var err error
	timeout := s.now.Add(time.Duration(s.SendTimeout) * time.Millisecond)
	err = cnx.SetWriteDeadline(timeout)
	if err != nil {
		goto errLabel
	}
	if _, err = cnx.Write(s.buffer[:s.bufferLen]); err != nil {
		goto errLabel
	}
	return

errLabel:
	s.connection.HandleError(err)
	s.connection.Disconnect()
	return
}
