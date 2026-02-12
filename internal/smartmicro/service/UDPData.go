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
	Buffer            [2000]byte `json:"-"`
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
	Metrics           udpDataMetrics
	utils.ErrorLoggerMixin
}

type udpDataMetrics struct {
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
	IsRunningMetric      *utils.Metric
}

func (u *udpDataMetrics) InitMetrics(serviceName string) {
	gm := &utils.GlobalMetrics
	sn := serviceName
	u.IsRunningMetric = gm.Metric(sn, "Is Running", utils.MetricTypeU32)
	u.SockOpenFailMetric = gm.Metric(sn, "Error: Socket Open Fail", utils.MetricTypeU64)
	u.SockWriteFailMetric = gm.Metric(sn, "Error: Socket WritePacket Fail", utils.MetricTypeU64)
	u.SockReadFailMetric = gm.Metric(sn, "Error: Socket Read Fail", utils.MetricTypeU64)
	u.SockReuseMetric = gm.Metric(sn, "Open Socket Reused", utils.MetricTypeU64)
	u.DataIterationsMetric = gm.Metric(sn, "Data Iterations", utils.MetricTypeU64)
	u.DataBytesMetric = gm.Metric(sn, "Bytes Received", utils.MetricTypeU64)
	u.NoDataMetric = gm.Metric(sn, "No Data Received", utils.MetricTypeU64)
	u.SocketSkipMetric = gm.Metric(sn, "Skip Failed Socket", utils.MetricTypeU64)
	u.SocketOpenMetric = gm.Metric(sn, "Socket Opens", utils.MetricTypeU64)
	u.UnmappedRadarPacket = gm.Metric(sn, "Error: Incorrect Radar", utils.MetricTypeU64)
	u.TimeMetric = gm.Metric(sn, "Now", utils.MetricTypeU64)
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
	u.Metrics.InitMetrics(u.GetServiceName())
	u.Terminate = false
	u.Terminated = false
	u.Metrics.IsRunningMetric.SetU32(1, time.Now())

	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)

	u.Connection.Init(u, u.ListenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, u.ReconnectCycle)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, context utils.IPErrorContext, err error) {
		switch context {
		case utils.IPErrorOnConnect:
			u.Metrics.SockOpenFailMetric.AddCount(1, u.Now)
		case utils.IPErrorOnWriteData:
			u.Metrics.SockWriteFailMetric.AddCount(1, u.Now)
		case utils.IPErrorOnReadData:
			u.Metrics.SockReadFailMetric.AddCount(1, u.Now)
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

	u.Metrics.IsRunningMetric.SetU32(0, time.Now())
	u.Terminated = true
}

func (u *UDPData) executeReader() {
	for !u.Terminate {
		u.Now = time.Now()
		u.Metrics.TimeMetric.SetTime(u.Now)

		if cnx := u.Connection.Listen(); cnx != nil {
			u.ClearError()
			u.Metrics.SockReuseMetric.AddCount(1, u.Now)

			u.BufferLen = u.Connection.ReceiveData(u.Buffer[:], u.Now, u.ReadTimeout)
			if u.BufferLen > 0 {
				u.Metrics.DataIterationsMetric.AddCount(1, u.Now)
				u.Metrics.DataBytesMetric.AddCount(uint64(u.BufferLen), u.Now)

				if u.OnData != nil {
					u.OnData(u, u.Connection.FromAddr, u.Buffer[:u.BufferLen])
				}
			} else {
				u.Metrics.NoDataMetric.AddCount(1, u.Now)
			}
		} else {
			u.Metrics.SocketSkipMetric.AddCount(1, u.Now)
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
	u.Metrics.SocketOpenMetric.AddCount(1, u.Now)
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
