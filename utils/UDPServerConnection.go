package utils

import (
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type UDPErrorContext uint8

const (
	UDPErrorOnConnect UDPErrorContext = iota
	UDPErrorOnWriteData
	UDPErrorOnReadData
)

func (conn *UDPErrorContext) String() string {
	switch *conn {
	case UDPErrorOnConnect:
		return "connect"
	case UDPErrorOnWriteData:
		return "send Data"
	case UDPErrorOnReadData:
		return "read Data"
	default:
		return "unknown"
	}
}

type UDPServerConnection struct {
	Sender          any
	OnError         func(*UDPServerConnection, UDPErrorContext, error)
	OnOpen          func(*UDPServerConnection)
	OnClose         func(*UDPServerConnection)
	address         net.UDPAddr
	connection      *net.UDPConn
	retry           RetryGuard
	readBufferSize  int
	writeBufferSize int
	FromAddr        net.UDPAddr
}

var ErrWriteToClosed = errors.New("write to closed connection")

// Init to initialize details for the connection
// sender is used indirectly from the callbacks.
// listenAddress is the UDP remoteAddr and port to bind to.
// readBufferSize is used in preparing the connection
func (m *UDPServerConnection) Init(
	sender any,
	listenAddress IP4,
	readBufferSize int,
	writeBufferSize int,
	reconnectOnCycle int,
) {
	m.address = listenAddress.ToUDPAddr()
	m.Sender = sender
	m.connection = nil
	m.readBufferSize = readBufferSize
	m.writeBufferSize = writeBufferSize
	m.retry = RetryGuard{
		RetryEvery: uint32(reconnectOnCycle),
	}
}

func (m *UDPServerConnection) Listen() bool {
	var err error

	if m.connection != nil {
		return true
	}

	if !m.retry.ShouldRetry() {
		return false
	}

	if m.connection, err = net.ListenUDP("udp4", &m.address); err != nil {
		goto errorLabel
	}

	if err = m.connection.SetReadBuffer(m.readBufferSize); err != nil {
		goto errorLabel
	}

	if err = m.connection.SetWriteBuffer(m.writeBufferSize); err != nil {
		goto errorLabel
	}

	m.retry.Reset()

	if m.OnOpen != nil {
		m.OnOpen(m)
	}
	return true

errorLabel:
	m.sendError(UDPErrorOnConnect, err)
	m.Close()
	return false
}

func (m *UDPServerConnection) ReceiveData(buffer []byte, now time.Time, waitMs int) (bufferLen int) {
	if m.connection == nil {
		return 0
	}
	var err error
	var fromAddr *net.UDPAddr

	deadline := now.Add(time.Duration(waitMs) * time.Millisecond)

	if err = m.connection.SetReadDeadline(deadline); err != nil {
		goto errorLabel
	}

	if bufferLen, fromAddr, err = m.connection.ReadFromUDP(buffer); err != nil {
		goto errorLabel
	}

	m.FromAddr = *fromAddr

	return

errorLabel:
	m.sendError(UDPErrorOnReadData, err)
	m.Close()
	return 0
}

func (m *UDPServerConnection) sendError(context UDPErrorContext, err error) {
	if m.OnError != nil {
		m.OnError(m, context, err)
	} else {
		log.Err(err).Str("context", context.String()).Msgf("UDPServerConnection")
	}
}

func (m *UDPServerConnection) Close() {
	if m.connection != nil {
		if m.OnClose != nil {
			m.OnClose(m)
		}
		_ = m.connection.Close()
		m.connection = nil
	}
}

func (m *UDPServerConnection) GetConnection() *net.UDPConn {
	return m.connection
}

func (m *UDPServerConnection) WriteData(udpAddr net.UDPAddr, buffer []byte) {
	var err error

	if m.connection == nil {
		m.sendError(UDPErrorOnWriteData, ErrWriteToClosed)
		return
	}

	if err = m.connection.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		goto errLabel
	}

	if _, err = m.connection.WriteToUDP(buffer, &udpAddr); err != nil {
		goto errLabel
	}

	return

errLabel:
	m.sendError(UDPErrorOnWriteData, err)
	return
}
