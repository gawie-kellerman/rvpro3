package service

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"rvpro3/radarvision.com/internal/general"
	"rvpro3/radarvision.com/utils"
)

const UDPDataServiceName = "UDP.Data.Service"
const udpDataEnabled = "udp.data.enabled"
const udpDataReadTimeout = "udp.data.read.timeout"
const udpDataReconnectCycle = "udp.data.reconnect.cycle"
const udpDataReconnectSleep = "udp.data.reconnect.sleep"
const udpDataLogRepeatMillis = "udp.data.log.repeat.millis"

var errKeepAliveNotFound = errors.New("keep alive not found")

type UDPDataService struct {
	Connection        utils.UDPServerConnection `json:"-"`
	Buffer            [4000]byte                `json:"-"`
	BufferLen         int                       `json:"-"`
	ListenAddr        utils.IP4
	Now               time.Time
	Terminate         bool
	Terminated        bool
	ReadTimeout       utils.Milliseconds
	ReconnectCycle    int
	ReconnectSleep    utils.Milliseconds
	IsEnabled         bool
	CurrentErr        utils.ErrorLoggerMixin
	Metrics           UdpDataMetrics
	TerminateRefCount atomic.Int32                               `json:"-"`
	OnData            func(*UDPDataService, net.UDPAddr, []byte) `json:"-"`
	OnError           func(*UDPDataService, error)               `json:"-"`
	OnTerminate       func(*UDPDataService)                      `json:"-"`
	doneChannel       chan bool
	writeChannel      chan UDPSendData
}

type UdpDataMetrics struct {
	ErrorsOnSocketConnect *utils.Metric
	ErrorsOnSocketWrite   *utils.Metric
	ErrorsOnSocketRead    *utils.Metric
	SocketReuseCount      *utils.Metric
	DataReadCount         *utils.Metric
	DataReadBytes         *utils.Metric
	NoDataCount           *utils.Metric
	SocketSkipCount       *utils.Metric
	SocketOpenCount       *utils.Metric
	InvalidRadarSkipCount *utils.Metric
	OnDataCallbackCount   *utils.Metric
	utils.MetricsInitMixin
}

func (u *UDPDataService) InitFromSettings(settings *utils.Settings) {
	u.IsEnabled = settings.Basic.GetBool(udpDataEnabled, true)
	u.ReadTimeout = settings.Basic.GetMilliseconds(udpDataReadTimeout, 3000)
	u.ReconnectCycle = settings.Basic.GetInt(udpDataReconnectCycle, 5)
	u.ReconnectSleep = settings.Basic.GetMilliseconds(udpDataReconnectSleep, 1000)
	u.CurrentErr.RepeatDuration = settings.Basic.GetMilliseconds(udpDataLogRepeatMillis, 60000)
}

func (u *UDPDataService) Start(state *utils.State, config *utils.Settings) {
	if !general.ServiceHelper.ShouldStart(state, config, u) {
		return
	}

	if !u.IsEnabled {
		return
	}

	keepAlive, ok := state.Get(UDPKeepAliveServiceName).(*UDPKeepAliveService)
	if !ok {
		u.CurrentErr.LogErrorAt(time.Now(), u.GetServiceName(), errKeepAliveNotFound)
		u.IsEnabled = false
		return
	}

	u.ListenAddr = keepAlive.LocalIPAddr
	u.init()

	go u.executeReader()
	go u.executeWriter()
}

func (u *UDPDataService) init() {
	u.Metrics.InitMetrics(u.GetServiceName(), &u.Metrics)
	u.Terminate = false
	u.Terminated = false

	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)

	u.Connection.Init(u, u.ListenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, u.ReconnectCycle)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, context utils.IPErrorContext, err error) {
		switch context {
		case utils.IPErrorOnConnect:
			u.Metrics.ErrorsOnSocketConnect.Inc(1)
		case utils.IPErrorOnWriteData:
			u.Metrics.ErrorsOnSocketWrite.Inc(1)
		case utils.IPErrorOnReadData:
			u.Metrics.ErrorsOnSocketRead.Inc(1)
		}
		u.sendError(err)
	}

	u.Connection.OnOpen = u.onOpenUDPSocket
}

func (u *UDPDataService) GetServiceName() string {
	return UDPDataServiceName
}

func (u *UDPDataService) WriteData(ip4 utils.IP4, data []byte) {
	if !u.Terminate {
		u.writeChannel <- UDPSendData{
			Address: ip4,
			Data:    data,
		}
	}
}

func (u *UDPDataService) Stop() {
	u.doneChannel <- true

	n := 0
	for u.TerminateRefCount.Load() > 0 && n < 10 {
		time.Sleep(100 * time.Millisecond)
		n = n + 1
	}

	u.Terminated = true
}

func (u *UDPDataService) executeReader() {
	for !u.Terminate {
		u.Now = time.Now()

		if cnx := u.Connection.Listen(); cnx != nil {
			u.CurrentErr.Clear()
			u.Metrics.SocketReuseCount.Inc(1)

			u.BufferLen = u.Connection.ReceiveData(u.Buffer[:], u.Now, u.ReadTimeout)
			if u.BufferLen > 0 {
				u.Metrics.DataReadCount.Inc(1)
				u.Metrics.DataReadBytes.Inc(int64(u.BufferLen))

				if u.OnData != nil {
					u.Metrics.OnDataCallbackCount.IncAt(1, u.Now)
					u.OnData(u, u.Connection.FromAddr, u.Buffer[:u.BufferLen])
				}
			} else {
				u.Metrics.NoDataCount.IncAt(1, u.Now)
			}
		} else {
			u.Metrics.SocketSkipCount.IncAt(1, u.Now)
			u.ReconnectSleep.Sleep()
		}
	}

	u.Connection.Close()
	u.TerminateRefCount.Add(-1)

	if u.OnTerminate != nil {
		u.OnTerminate(u)
	}
}

func (u *UDPDataService) onOpenUDPSocket(_ *utils.UDPServerConnection) {
	u.Metrics.SocketOpenCount.Inc(1)
}

func (u *UDPDataService) sendError(err error) {
	if u.OnError != nil {
		u.OnError(u, err)
	} else {
		u.CurrentErr.LogErrorAt(time.Now(), u.GetServiceName(), err)
	}
}

func (u *UDPDataService) executeWriter() {
	for {
		select {
		case data := <-u.writeChannel:
			u.writeData(data)

		case <-u.doneChannel:
			u.TerminateRefCount.Add(-1)
			u.Terminate = true
			close(u.doneChannel)
			close(u.writeChannel)
			return
		}
	}
}

func (u *UDPDataService) writeData(data UDPSendData) {
	udpAddr := data.Address.ToUDPAddr()
	if err := u.Connection.WriteData(udpAddr, data.Data); err != nil {
		// TODO: Not sure how this is going to play out
		u.CurrentErr.LogErrorAt(time.Now(), u.GetServiceName(), err)
	}
}

func (u *UDPDataService) IsTerminated() bool {
	return u.Terminated
}

type UDPSendData struct {
	Address utils.IP4
	Data    []byte
}
