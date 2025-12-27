package uartsdlc

import (
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
	"rvpro3/radarvision.com/utils"
)

var errConnectionClosed = errors.New("serial connection closed")

type SerialConnection struct {
	PortName     string
	connection   serial.Port
	mode         serial.Mode
	terminate    bool
	RetryGuard   utils.RetryGuard
	OnConnect    func(*SerialConnection)
	OnDisconnect func(*SerialConnection)
	OnError      func(*SerialConnection, error)
	OnWrote      func(*SerialConnection, []byte)
	OnRead       func(*SerialConnection, []byte)
}

func (s *SerialConnection) Init(
	portName string,
	baudRate int,
	dataBits int,
	parity serial.Parity,
	stopBits serial.StopBits,
) {
	s.PortName = portName
	s.mode.Parity = parity
	s.mode.StopBits = stopBits
	s.mode.BaudRate = baudRate
	s.mode.DataBits = dataBits
	s.RetryGuard.RetryEvery = 3
	s.terminate = false
}

func (s *SerialConnection) Connect() bool {
	var err error

	if s.connection != nil {
		return true
	}

	if !s.RetryGuard.ShouldRetry() {
		return false
	}

	if s.connection, err = serial.Open(s.PortName, &s.mode); err != nil {
		goto errorLabel
	}

	s.RetryGuard.Reset()

	if s.OnConnect != nil {
		s.OnConnect(s)
	}
	return true

errorLabel:
	s.HandleError(err)
	s.Disconnect()
	return false
}

func (s *SerialConnection) Read(buffer []byte) int {
	var bytesRead int
	var err error

	_ = s.connection.SetReadTimeout(1 * time.Second)

	if bytesRead, err = s.connection.Read(buffer); err != nil {
		if err != io.EOF {
			s.Disconnect()
			s.HandleError(err)
		}
		return 0
	}

	if s.OnRead != nil {
		s.OnRead(s, buffer[:bytesRead])
	}
	return bytesRead
}

func (s *SerialConnection) HandleError(err error) {
	if s.OnError != nil {
		s.OnError(s, err)
		log.Err(err).Msg("Serial connection error")
	} else {
		log.Err(err).Msg("Serial connection error")
	}
}

func (s *SerialConnection) Disconnect() {
	if s.connection != nil {
		if s.OnDisconnect != nil {
			s.OnDisconnect(s)
		}
		_ = s.connection.Close()
		s.connection = nil
	}
}

func (s *SerialConnection) Write(data []byte) bool {
	if s.connection != nil {
		if _, err := s.connection.Write(data); err != nil {
			s.Disconnect()
			s.HandleError(err)
			return false
		}

		if err := s.connection.Drain(); err != nil {
			s.HandleError(err)
			return false
		}
		if s.OnWrote != nil {
			s.OnWrote(s, data)
		}
	} else {
		s.HandleError(errConnectionClosed)
		return false
	}
	return true
}
