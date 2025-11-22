package utils

import (
	"net"

	"github.com/rs/zerolog/log"
)

type UDPClientConnection struct {
	Owner           any
	connection      *net.UDPConn
	retryGuard      RetryGuard
	LocalIPAddr     IP4
	MulticastIPAddr IP4
	OnConnect       func(*UDPClientConnection)
	OnDisconnect    func(*UDPClientConnection)
	OnError         func(*UDPClientConnection, error)
}

func (s *UDPClientConnection) Init(
	localIPAddr IP4,
	multicastIPAddr IP4,
	owner any,
	reconnectOnCycle int,
) {
	s.LocalIPAddr = localIPAddr.WithPort(0)
	s.MulticastIPAddr = multicastIPAddr
	s.Owner = owner
	s.retryGuard.RetryEvery = uint32(reconnectOnCycle)
}

func (s *UDPClientConnection) GetConnection() *net.UDPConn {
	return s.connection
}

func (s *UDPClientConnection) Connect() bool {
	var err error

	if s.connection != nil {
		return true
	}

	if !s.retryGuard.ShouldRetry() {
		return false
	}

	mAddr := s.MulticastIPAddr.ToUDPAddr()
	lAddr := s.LocalIPAddr.ToUDPAddr()

	if s.connection, err = net.DialUDP("udp4", &lAddr, &mAddr); err != nil {
		goto errorLabel
	}

	s.retryGuard.Reset()
	if s.OnConnect != nil {
		s.OnConnect(s)
	}
	return true

errorLabel:
	s.HandleError(err)
	return false
}

func (s *UDPClientConnection) Disconnect() {
	if s.connection != nil {
		if s.OnDisconnect != nil {
			s.OnDisconnect(s)
		}
		_ = s.connection.Close()
		s.connection = nil
	}
}

func (s *UDPClientConnection) HandleError(err error) {
	if s.OnError != nil {
		s.OnError(s, err)
	} else {
		log.Err(err).Msg("UDPClientConnection")
	}
}
