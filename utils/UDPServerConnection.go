package utils

import "net"

type UDPServerConnection struct {
	Sender          any
	OnError         func(*UDPServerConnection, error)
	OnOpen          func(*UDPServerConnection)
	OnClose         func(*UDPServerConnection)
	address         net.UDPAddr
	connection      *net.UDPConn
	retry           RetryGuard
	readBufferSize  int
	writeBufferSize int
}

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
		ModCycles: uint32(reconnectOnCycle),
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
	if m.OnError != nil {
		m.OnError(m, err)
	}
	m.Close()
	return false
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
