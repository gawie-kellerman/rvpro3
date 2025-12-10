package service

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

type UDPDataService struct {
	Connection        utils.UDPServerConnection
	OnData            func(*UDPDataService, net.UDPAddr, []byte)
	OnError           func(*UDPDataService, error)
	OnTerminate       func(*UDPDataService)
	Buffer            [udpBufferSize]byte
	BufferLen         int
	ListenAddr        utils.IP4
	Now               time.Time
	doneChannel       chan bool
	writeChannel      chan UDPSendData
	terminate         bool
	TerminateRefCount atomic.Int32
	Stats             UDPDataServiceStatistics
}

func (u *UDPDataService) WriteData(ip4 utils.IP4, data []byte) {
	if !u.terminate {
		u.writeChannel <- UDPSendData{
			Address: ip4,
			Data:    data,
		}
	}
}

func (u *UDPDataService) Start(listenAddr utils.IP4) {
	u.Stats.Init()
	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)
	u.ListenAddr = listenAddr
	u.Connection.Init(u, listenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, 3)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, err error) {
		u.sendError(err)
	}
	go u.executeReader()
	go u.executeWriter()
}

func (u *UDPDataService) Stop() {
	u.doneChannel <- true

	for u.TerminateRefCount.Load() > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (u *UDPDataService) executeReader() {
	for !u.terminate {
		u.Now = time.Now()
		if u.Connection.Listen() {
			u.Stats.Register(udpSocketSuccess, u.Now)

			u.BufferLen = u.Connection.ReceiveData(u.Buffer[:], u.Now, 3000)
			if u.BufferLen > 0 {
				u.Stats.Register(udpDataReceived, u.Now)
				u.Stats.Aggregate(udpDataTotal, uint64(u.BufferLen), u.Now)
				if u.OnData != nil {
					u.OnData(u, u.Connection.FromAddr, u.Buffer[:u.BufferLen])
				}
			} else {
				u.Stats.Register(updDataNotReceived, u.Now)
			}
		} else {
			u.Stats.Register(updSocketFailure, u.Now)
		}
	}

	u.Connection.Close()
	u.TerminateRefCount.Add(-1)

	if u.OnTerminate != nil {
		u.OnTerminate(u)
	}
}

func (u *UDPDataService) sendError(err error) {
	if u.OnError != nil {
		u.OnError(u, err)
	} else {
		log.Err(err).Msgf("UDPDataService")
	}
}

func (u *UDPDataService) executeWriter() {
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

func (u *UDPDataService) writeData(data UDPSendData) {
	udpAddr := data.Address.ToUDPAddr()
	u.Connection.WriteData(udpAddr, data.Data)
}

type UDPSendData struct {
	Address utils.IP4
	Data    []byte
}
