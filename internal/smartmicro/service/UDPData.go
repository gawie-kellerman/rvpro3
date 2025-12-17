package service

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/instrumentation"
	"rvpro3/radarvision.com/utils"
)

type UDPData struct {
	Metrics           *instrumentation.Metrics
	Connection        utils.UDPServerConnection
	Buffer            [udpBufferSize]byte
	BufferLen         int
	ListenAddr        utils.IP4
	Now               time.Time
	OnData            func(*UDPData, net.UDPAddr, []byte)
	OnError           func(*UDPData, error)
	OnTerminate       func(*UDPData)
	doneChannel       chan bool
	writeChannel      chan UDPSendData
	terminate         bool
	TerminateRefCount atomic.Int32
}

func (u *UDPData) WriteData(ip4 utils.IP4, data []byte) {
	if !u.terminate {
		u.writeChannel <- UDPSendData{
			Address: ip4,
			Data:    data,
		}
	}
}

func (u *UDPData) Start(listenAddr utils.IP4) {
	u.Metrics = &instrumentation.GlobalUDPMetrics
	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)
	u.ListenAddr = listenAddr

	u.Connection.Init(u, listenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, 3)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, context utils.UDPErrorContext, err error) {
		switch context {
		case utils.UDPErrorOnConnect:
			u.CountMetric(instrumentation.UDPMetricSocketOpenFail, 1)
		case utils.UDPErrorOnWriteData:
			u.CountMetric(instrumentation.UDPMetricSocketWriteFail, 1)
		case utils.UDPErrorOnReadData:
			u.CountMetric(instrumentation.UDPMetricSocketReadFail, 1)
		}
		u.sendError(err)
	}

	u.Connection.OnOpen = u.onOpenUDPSocket

	go u.executeReader()
	go u.executeWriter()
}

func (u *UDPData) Stop() {
	u.doneChannel <- true

	for u.TerminateRefCount.Load() > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (u *UDPData) executeReader() {
	for !u.terminate {
		u.Now = time.Now()

		// Record the current time in milliseconds
		u.Metrics.SetTime(
			int(instrumentation.UDPNow),
			u.Now,
		)

		if u.Connection.Listen() {
			u.CountMetric(instrumentation.UDPMetricSocketUse, 1)

			u.BufferLen = u.Connection.ReceiveData(u.Buffer[:], u.Now, 3000)
			if u.BufferLen > 0 {
				//
				u.CountMetric(instrumentation.UDPMetricDataIterations, 1)
				u.CountMetric(instrumentation.UDPMetricDataBytes, uint64(u.BufferLen))

				if u.OnData != nil {
					u.OnData(u, u.Connection.FromAddr, u.Buffer[:u.BufferLen])
				}
			} else {
				u.CountMetric(instrumentation.UDPMetricNoDataReceived, 1)
			}
		} else {
			u.CountMetric(instrumentation.UDPMetricSocketSkip, 1)
		}
	}

	u.Connection.Close()
	u.TerminateRefCount.Add(-1)

	if u.OnTerminate != nil {
		u.OnTerminate(u)
	}
}

func (u *UDPData) CountMetric(metric instrumentation.UDPMetric, count uint64) {
	u.Metrics.AddCount(int(metric), count, u.Now)
}

func (u *UDPData) onOpenUDPSocket(connection *utils.UDPServerConnection) {
	u.CountMetric(instrumentation.UDPMetricSocketOpen, 1)
}

func (u *UDPData) sendError(err error) {
	if u.OnError != nil {
		u.OnError(u, err)
	} else {
		log.Err(err).Msgf("UDPData")
	}
}

func (u *UDPData) executeWriter() {
	for {
		select {
		case data := <-u.writeChannel:
			u.writeData(data)

		case <-u.doneChannel:
			u.TerminateRefCount.Add(-1)
			u.terminate = true
			close(u.doneChannel)
			close(u.writeChannel)
			return
		}
	}
}

func (u *UDPData) writeData(data UDPSendData) {
	udpAddr := data.Address.ToUDPAddr()
	u.Connection.WriteData(udpAddr, data.Data)
}

type UDPSendData struct {
	Address utils.IP4
	Data    []byte
}
