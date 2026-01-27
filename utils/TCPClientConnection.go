package utils

import (
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type TCPClientConnection struct {
	Owner           any
	OnError         func(*TCPClientConnection, IPErrorContext, error)
	OnConnect       func(*TCPClientConnection)
	OnDisconnect    func(*TCPClientConnection)
	remoteAddr      net.TCPAddr
	localAddr       *net.TCPAddr
	connection      *net.TCPConn
	retry           RetryGuard
	readBufferSize  int
	writeBufferSize int
}

func (cm *TCPClientConnection) Init(
	owner any,
	remoteAddr IP4,
	reconnectOnCycle int,
	readBufferSize int,
	writeBufferSize int,
) {
	cm.Owner = owner
	cm.remoteAddr = remoteAddr.ToTCPAddr()
	cm.connection = nil
	cm.retry = RetryGuard{
		RetryEvery: uint32(reconnectOnCycle),
	}
	cm.readBufferSize = readBufferSize
	cm.writeBufferSize = writeBufferSize
}

func (cm *TCPClientConnection) GetConnection() *net.TCPConn {
	if cm.Connect() {
		return cm.connection
	}
	return nil
}

// SetLocalAddr must be called before opening/reopening the connection
func (cm *TCPClientConnection) SetLocalAddr(localAddr *net.TCPAddr) {
	cm.localAddr = localAddr
}

func (cm *TCPClientConnection) Read(buffer []byte, now time.Time, waitMs int) (bufferLen int) {
	var err error
	cnx := cm.GetConnection()

	if cnx == nil {
		return
	}

	deadline := now.Add(time.Duration(waitMs) * time.Millisecond)

	if err = cnx.SetReadDeadline(deadline); err != nil {
		goto errorLabel
	}

	bufferLen, err = cnx.Read(buffer)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return 0
		}

		goto errorLabel
	}

	return bufferLen

errorLabel:
	cm.sendError(IPErrorOnReadData, err)
	cm.Close()
	return 0
}

func (cm *TCPClientConnection) Write(buffer []byte, now time.Time, waitMs int) bool {
	var err error
	cnx := cm.GetConnection()

	if cnx == nil {
		return false
	}

	deadline := now.Add(time.Duration(waitMs) * time.Millisecond)
	//
	if err = cnx.SetWriteDeadline(deadline); err != nil {
		goto errorLabel
	}

	if _, err = cnx.Write(buffer); err != nil {
		goto errorLabel
	}

	return true

errorLabel:
	cm.sendError(IPErrorOnWriteData, err)
	cm.Close()
	return false
}

func (cm *TCPClientConnection) Close() {
	if cm.connection != nil {
		_ = cm.connection.Close()
		cm.connection = nil
	}
}

func (cm *TCPClientConnection) sendError(context IPErrorContext, err error) {
	if cm.OnError != nil {
		cm.OnError(cm, context, err)
	}
}

func (cm *TCPClientConnection) Connect() bool {
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
	if cm.OnConnect != nil {
		cm.OnConnect(cm)
	}
	return true

errorLabel:
	cm.onError(IPErrorOnConnect, err)
	cm.Disconnect()
	return false
}

func (cm *TCPClientConnection) Disconnect() {
	if cm.connection != nil {
		if cm.OnDisconnect != nil {
			cm.OnDisconnect(cm)
		}
		_ = cm.connection.Close()
		cm.connection = nil
	}
}

func (cm *TCPClientConnection) onError(context IPErrorContext, err error) {
	if cm.OnError != nil {
		cm.OnError(cm, context, err)
	} else {
		log.Err(err).Msg("TCPClientConnection.HandleError")
	}
}
