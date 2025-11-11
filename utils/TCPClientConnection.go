package utils

import (
	"net"

	"github.com/rs/zerolog/log"
)

type TCPClientConnection struct {
	Sender          any
	OnError         func(*TCPClientConnection, error)
	OnOpen          func(*TCPClientConnection)
	OnClose         func(*TCPClientConnection)
	remoteAddr      net.TCPAddr
	localAddr       *net.TCPAddr
	connection      *net.TCPConn
	retry           RetryGuard
	readBufferSize  int
	writeBufferSize int
}

func (cm *TCPClientConnection) Init(
	sender any,
	remoteAddr IP4,
	reconnectOnCycle int,
	readBufferSize int,
	writeBufferSize int,
) {
	cm.Sender = sender
	cm.remoteAddr = remoteAddr.ToTCPAddr()
	cm.connection = nil
	cm.retry = RetryGuard{
		ModCycles: uint32(reconnectOnCycle),
	}
	cm.readBufferSize = readBufferSize
	cm.writeBufferSize = writeBufferSize
}

func (cm *TCPClientConnection) GetConnection() *net.TCPConn {
	return cm.connection
}

// SetLocalAddr must be called before opening/reopening the connection
func (cm *TCPClientConnection) SetLocalAddr(localAddr *net.TCPAddr) {
	cm.localAddr = localAddr
}

func (cm *TCPClientConnection) Open() bool {
	var err error

	if cm.connection != nil {
		return true
	}

	if !cm.retry.ShouldRetry() {
		return false
	}

	if cm.connection, err = net.DialTCP("tcp", nil, &cm.remoteAddr); err != nil {
		goto errorLabel
	}

	if err = cm.connection.SetReadBuffer(cm.readBufferSize); err != nil {
		goto errorLabel
	}

	if err = cm.connection.SetWriteBuffer(cm.writeBufferSize); err != nil {
		goto errorLabel
	}

	cm.retry.Reset()
	if cm.OnOpen != nil {
		cm.OnOpen(cm)
	}
	return true

errorLabel:
	cm.onError(err)

	cm.Close()
	return false
}

func (cm *TCPClientConnection) Close() {
	if cm.connection != nil {
		if cm.OnClose != nil {
			cm.OnClose(cm)
		}
		_ = cm.connection.Close()
		cm.connection = nil
	}
}

func (cm *TCPClientConnection) onError(err error) {
	if cm.OnError != nil {
		cm.OnError(cm, err)
	} else {
		log.Err(err).Msg("TCPClientConnection.onError")
	}
}
