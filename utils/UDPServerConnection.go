package utils

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type IPErrorContext uint8

const (
	IPErrorOnConnect IPErrorContext = iota
	IPErrorOnWriteData
	IPErrorOnReadData
)

func (conn *IPErrorContext) String() string {
	switch *conn {
	case IPErrorOnConnect:
		return "connect"
	case IPErrorOnWriteData:
		return "send Metric"
	case IPErrorOnReadData:
		return "read Metric"
	default:
		return "unknown"
	}
}

type UDPServerConnection struct {
	Sender          any                                               `json:"-"`
	OnError         func(*UDPServerConnection, IPErrorContext, error) `json:"-"`
	OnOpen          func(*UDPServerConnection)                        `json:"-"`
	OnClose         func(*UDPServerConnection)                        `json:"-"`
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
	m.retry = RetryGuard{
		RetryEvery: uint32(reconnectOnCycle),
	}
	m.writeBufferSize = writeBufferSize
}

func (m *UDPServerConnection) Listen() *net.UDPConn {
	var err error

	if m.connection != nil {
		return m.connection
	}

	if !m.retry.ShouldRetry() {
		return nil
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
	return m.connection

errorLabel:
	m.sendError(IPErrorOnConnect, err)
	m.Close()
	return m.connection
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
	m.sendError(IPErrorOnReadData, err)
	m.Close()
	return 0
}

func (m *UDPServerConnection) Read(buffer []byte, now time.Time, waitMs int) (bufferLen int, fromAddr IP4) {
	cnx := m.Listen()

	if cnx == nil {
		return 0, fromAddr
	}

	var err error
	var udpAddr *net.UDPAddr

	deadline := now.Add(time.Duration(waitMs) * time.Millisecond)

	if err = cnx.SetReadDeadline(deadline); err != nil {
		goto errorLabel
	}

	if bufferLen, udpAddr, err = cnx.ReadFromUDP(buffer); err != nil {
		// If no data received then handle it gracefully
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return 0, fromAddr
		}
		goto errorLabel
	}
	fromAddr = IP4Builder.FromAddr(udpAddr)

	return bufferLen, fromAddr

errorLabel:
	m.sendError(IPErrorOnReadData, err)
	m.Close()
	return 0, fromAddr
}

func (m *UDPServerConnection) sendError(context IPErrorContext, err error) {
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
	m.Listen()
	return m.connection
}

func (m *UDPServerConnection) WriteData(udpAddr net.UDPAddr, buffer []byte) error {
	var err error

	cnx := m.GetConnection()

	if cnx == nil {
		err = ErrWriteToClosed
		goto errLabel
	}

	if err = cnx.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		goto errLabel
	}

	fmt.Println("Writing from", cnx.LocalAddr(), "to", udpAddr.String(), len(buffer), "bytes")
	if _, err = cnx.WriteToUDP(buffer, &udpAddr); err != nil {
		goto errLabel
	}

	return nil

errLabel:
	m.sendError(IPErrorOnWriteData, err)
	return err
}
