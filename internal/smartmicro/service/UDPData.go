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
	Connection           utils.UDPServerConnection
	Buffer               [1500]byte `json:"-"`
	BufferLen            int
	ListenAddr           utils.IP4
	Now                  time.Time
	OnData               func(*UDPData, net.UDPAddr, []byte) `json:"-"`
	OnError              func(*UDPData, error)               `json:"-"`
	OnTerminate          func(*UDPData)                      `json:"-"`
	doneChannel          chan bool
	writeChannel         chan UDPSendData
	Terminate            bool
	Terminated           bool
	TerminateRefCount    atomic.Int32 `json:"-"`
	ReadTimeout          int
	ReconnectCycle       int
	ReconnectSleep       int
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
	IncorrectRadarMetric *utils.Metric
	IsRunningMetric      *utils.Metric
	utils.ErrorLoggerMixin
}

func (u *UDPData) SetupDefaults(config *utils.Config) {
	config.SetSettingAsBool(udpDataEnabled, true)
	config.SetSettingAsInt(udpDataReadTimeout, 3000)
	config.SetSettingAsInt(udpDataReconnectCycle, 5)
	config.SetSettingAsMillis(udpDataReconnectSleep, 1000)
	config.SetSettingAsInt(udpDataLogRepeatMillis, 60000) // only log continuous failures once a minute
}

func (u *UDPData) SetupRunnable(state *utils.State, config *utils.Config) {
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

func (u *UDPData) Init() {
	u.ListenAddr = utils.IP4Builder.FromString("192.168.11.2:55555")
	u.ReadTimeout = 3000
	u.ReconnectCycle = 3
	u.ReconnectSleep = 1000
}

func (u *UDPData) InitMetrics() {
	gm := &utils.GlobalMetrics
	sn := u.GetServiceName()
	u.IsRunningMetric = gm.Metric(sn, "Is Running", utils.MetricTypeU32)
	u.SockOpenFailMetric = gm.Metric(sn, "Error: Socket Open Fail", utils.MetricTypeU64)
	u.SockWriteFailMetric = gm.Metric(sn, "Error: Socket Write Fail", utils.MetricTypeU64)
	u.SockReadFailMetric = gm.Metric(sn, "Error: Socket Read Fail", utils.MetricTypeU64)
	u.SockReuseMetric = gm.Metric(sn, "Open Socket Reused", utils.MetricTypeU64)
	u.DataIterationsMetric = gm.Metric(sn, "Metric Iterations", utils.MetricTypeU64)
	u.DataBytesMetric = gm.Metric(sn, "Bytes Received", utils.MetricTypeU64)
	u.NoDataMetric = gm.Metric(sn, "No Metric Received", utils.MetricTypeU64)
	u.SocketSkipMetric = gm.Metric(sn, "Skip Failed Socket", utils.MetricTypeU64)
	u.SocketOpenMetric = gm.Metric(sn, "Socket Opens", utils.MetricTypeU64)
	u.IncorrectRadarMetric = gm.Metric(sn, "Error: Incorrect Radar", utils.MetricTypeU64)
	u.TimeMetric = gm.Metric(sn, "Now", utils.MetricTypeU64)
}

func (u *UDPData) Start() {
	u.InitMetrics()
	u.Terminate = false
	u.Terminated = false
	u.IsRunningMetric.SetU32(1, time.Now())

	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)

	u.Connection.Init(u, u.ListenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, u.ReconnectCycle)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, context utils.UDPErrorContext, err error) {
		switch context {
		case utils.UDPErrorOnConnect:
			u.SockOpenFailMetric.AddCount(1, u.Now)
		case utils.UDPErrorOnWriteData:
			u.SockWriteFailMetric.AddCount(1, u.Now)
		case utils.UDPErrorOnReadData:
			u.SockReadFailMetric.AddCount(1, u.Now)
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

	u.IsRunningMetric.SetU32(0, time.Now())
	u.Terminated = true
}

func (u *UDPData) executeReader() {
	for !u.Terminate {
		u.Now = time.Now()
		u.TimeMetric.SetTime(u.Now)

		if u.Connection.Listen() {
			u.ClearError()
			u.SockReuseMetric.AddCount(1, u.Now)

			u.BufferLen = u.Connection.ReceiveData(u.Buffer[:], u.Now, u.ReadTimeout)
			if u.BufferLen > 0 {
				u.DataIterationsMetric.AddCount(1, u.Now)
				u.DataBytesMetric.AddCount(uint64(u.BufferLen), u.Now)

				if u.OnData != nil {
					u.OnData(u, u.Connection.FromAddr, u.Buffer[:u.BufferLen])
				}
			} else {
				u.NoDataMetric.AddCount(1, u.Now)
			}
		} else {
			u.SocketSkipMetric.AddCount(1, u.Now)
			time.Sleep(time.Duration(u.ReconnectSleep) * time.Millisecond)
		}
	}

	u.Connection.Close()
	u.TerminateRefCount.Add(-1)

	if u.OnTerminate != nil {
		u.OnTerminate(u)
	}
}

func (u *UDPData) onOpenUDPSocket(connection *utils.UDPServerConnection) {
	u.SocketOpenMetric.AddCount(1, u.Now)
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
	u.Connection.WriteData(udpAddr, data.Data)
}

func (u *UDPData) IsTerminated() bool {
	return u.Terminated
}

func (u *UDPData) InitFromConfig(config *utils.Config) {
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
