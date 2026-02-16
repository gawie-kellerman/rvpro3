package service

import (
	"net"
	"sync/atomic"
	"time"

	"rvpro3/radarvision.com/utils"
)

const UDPDataServiceName = "UDP.Data.Service"
const udpDataEnabled = "UDP.Data.Enabled"
const udpDataReadTimeout = "UDP.Data.ReadTimeout"
const udpDataReconnectCycle = "UDP.Data.Reconnect.Cycle"
const udpDataReconnectSleep = "UDP.Data.Reconnect.AwaitClick"
const udpDataLogRepeatMillis = "UDP.Data.Log.RepeatMillis"

type UDPData struct {
	Connection        utils.UDPServerConnection
	Buffer            [4000]byte `json:"-"`
	BufferLen         int
	ListenAddr        utils.IP4
	Now               time.Time
	OnData            func(*UDPData, net.UDPAddr, []byte) `json:"-"`
	OnError           func(*UDPData, error)               `json:"-"`
	OnTerminate       func(*UDPData)                      `json:"-"`
	doneChannel       chan bool
	writeChannel      chan UDPSendData
	Terminate         bool
	Terminated        bool
	TerminateRefCount atomic.Int32 `json:"-"`
	ReadTimeout       int
	ReconnectCycle    int
	ReconnectSleep    int
	Metrics           UdpDataMetrics
	utils.ErrorLoggerMixin
}

type UdpDataMetrics struct {
	TimeMetric           *utils.Metric
	SockOpenFailMetric   *utils.Metric
	SockWriteFailMetric  *utils.Metric
	SockReadFailMetric   *utils.Metric
	SockReuseMetric      *utils.Metric
	DataIterationsMetric *utils.Metric
	DataBytesMetric      *utils.Metric
	NoDataMetric         *utils.Metric
	SocketSkipMetric     *utils.Metric
	SocketOpenMetric     *utils.Metric
	UnmappedRadarPacket  *utils.Metric
	IsRunningTime        *utils.Metric
	utils.MetricsInitMixin
}

func (u *UDPData) SetupDefaults(config *utils.Settings) {
	config.SetSettingAsBool(udpDataEnabled, true)
	config.SetSettingAsInt(udpDataReadTimeout, 3000)
	config.SetSettingAsInt(udpDataReconnectCycle, 5)
	config.SetSettingAsMillis(udpDataReconnectSleep, 1000)
	config.SetSettingAsInt(udpDataLogRepeatMillis, 60000) // only log continuous failures once a minute
}

func (u *UDPData) SetupAndStart(state *utils.State, config *utils.Settings) {
	if !config.GetSettingAsBool(udpDataEnabled) {
		return
	}

	u.InitFromConfig(config)
	u.Start()

	state.Set(u.GetServiceName(), u)
}

func (u *UDPData) GetServiceName() string {
	return UDPDataServiceName
}

func (u *UDPData) GetServiceNames() []string {
	return nil
}

func (u *UDPData) WriteData(ip4 utils.IP4, data []byte) {
	if !u.Terminate {
		u.writeChannel <- UDPSendData{
			Address: ip4,
			Data:    data,
		}
	}
}

func (u *UDPData) Start() {
	u.Metrics.InitMetrics(u.GetServiceName(), &u.Metrics)
	u.Terminate = false
	u.Terminated = false
	u.Metrics.IsRunningTime.SetTime()

	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)

	u.Connection.Init(u, u.ListenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, u.ReconnectCycle)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, context utils.IPErrorContext, err error) {
		switch context {
		case utils.IPErrorOnConnect:
			u.Metrics.SockOpenFailMetric.Inc(1)
		case utils.IPErrorOnWriteData:
			u.Metrics.SockWriteFailMetric.Inc(1)
		case utils.IPErrorOnReadData:
			u.Metrics.SockReadFailMetric.Inc(1)
		}
		u.sendError(err)
	}

	u.Connection.OnOpen = u.onOpenUDPSocket

	go u.executeReader()
	go u.executeWriter()
}

func (u *UDPData) Stop() {
	u.doneChannel <- true

	n := 0
	for u.TerminateRefCount.Load() > 0 && n < 10 {
		time.Sleep(100 * time.Millisecond)
		n = n + 1
	}

	u.Metrics.IsRunningTime.SetTime()
	u.Terminated = true
}

func (u *UDPData) executeReader() {
	for !u.Terminate {
		u.Now = time.Now()
		u.Metrics.TimeMetric.SetTime()

		if cnx := u.Connection.Listen(); cnx != nil {
			u.ClearError()
			u.Metrics.SockReuseMetric.Inc(1)

			u.BufferLen = u.Connection.ReceiveData(u.Buffer[:], u.Now, u.ReadTimeout)
			if u.BufferLen > 0 {
				u.Metrics.DataIterationsMetric.Inc(1)
				u.Metrics.DataBytesMetric.Inc(int64(u.BufferLen))

				if u.OnData != nil {
					u.OnData(u, u.Connection.FromAddr, u.Buffer[:u.BufferLen])
				}
			} else {
				u.Metrics.NoDataMetric.Inc(1)
			}
		} else {
			u.Metrics.SocketSkipMetric.Inc(1)
			time.Sleep(time.Duration(u.ReconnectSleep) * time.Millisecond)
		}
	}

	u.Connection.Close()
	u.TerminateRefCount.Add(-1)

	if u.OnTerminate != nil {
		u.OnTerminate(u)
	}
}

func (u *UDPData) onOpenUDPSocket(_ *utils.UDPServerConnection) {
	u.Metrics.SocketOpenMetric.Inc(1)
}

func (u *UDPData) sendError(err error) {
	if u.OnError != nil {
		u.OnError(u, err)
	} else {
		u.LogError("UDPData", err)
	}
}

func (u *UDPData) executeWriter() {
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

func (u *UDPData) writeData(data UDPSendData) {
	udpAddr := data.Address.ToUDPAddr()
	if err := u.Connection.WriteData(udpAddr, data.Data); err != nil {
		// TODO: Not sure how this is going to play out
		u.LogError("writeData", err)
	}
}

func (u *UDPData) IsTerminated() bool {
	return u.Terminated
}

func (u *UDPData) InitFromConfig(config *utils.Settings) {
	u.ListenAddr = config.GetSettingAsIP(udpKeepAliveCallbackIP)
	u.ReadTimeout = config.GetSettingAsInt(udpDataReadTimeout)
	u.ReconnectCycle = config.GetSettingAsInt(udpDataReconnectCycle)
	u.ReconnectSleep = config.GetSettingAsInt(udpDataReconnectSleep)
	u.LogRepeatMillis = config.GetSettingAsMillis(udpDataLogRepeatMillis)
}

type UDPSendData struct {
	Address utils.IP4
	Data    []byte
}
