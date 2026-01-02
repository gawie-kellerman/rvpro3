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
	MetricsAt            string
	Connection           utils.UDPServerConnection
	Buffer               [1500]byte
	BufferLen            int
	ListenAddr           utils.IP4
	Now                  time.Time
	OnData               func(*UDPData, net.UDPAddr, []byte)
	OnError              func(*UDPData, error)
	OnTerminate          func(*UDPData)
	doneChannel          chan bool
	writeChannel         chan UDPSendData
	terminate            bool
	TerminateRefCount    atomic.Int32
	ReadTimeout          int
	ReconnectCycle       int
	ReconnectSleep       int
	timeMetric           *utils.Metric
	sockOpenFailMetric   *utils.Metric
	sockWriteFailMetric  *utils.Metric
	sockReadFailMetric   *utils.Metric
	sockReuseMetric      *utils.Metric
	dataIterationsMetric *utils.Metric
	dataBytesMetric      *utils.Metric
	noDataMetric         *utils.Metric
	socketSkipMetric     *utils.Metric
	socketOpenMetric     *utils.Metric
	utils.ErrorLoggerMixin
	IncorrectRadarMetric *utils.Metric
	isRunningMetric      *utils.Metric
	terminated           bool
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

	state.Set(u.GetStateName(), u)
}

func (u *UDPData) GetStateName() string {
	return UDPDataServiceName
}

func (u *UDPData) GetStateNames() []string {
	return nil
}

func (u *UDPData) WriteData(ip4 utils.IP4, data []byte) {
	if !u.terminate {
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
	u.MetricsAt = "UDP.Data"
	gm := &utils.GlobalMetrics
	u.isRunningMetric = gm.Metric(u.MetricsAt, "Is Running", utils.MetricTypeU32)
	u.sockOpenFailMetric = gm.Metric(u.MetricsAt, "Error: Socket Open Fail", utils.MetricTypeU64)
	u.sockWriteFailMetric = gm.Metric(u.MetricsAt, "Error: Socket Write Fail", utils.MetricTypeU64)
	u.sockReadFailMetric = gm.Metric(u.MetricsAt, "Error: Socket Read Fail", utils.MetricTypeU64)
	u.sockReuseMetric = gm.Metric(u.MetricsAt, "Open Socket Reused", utils.MetricTypeU64)
	u.dataIterationsMetric = gm.Metric(u.MetricsAt, "Metric Iterations", utils.MetricTypeU64)
	u.dataBytesMetric = gm.Metric(u.MetricsAt, "Bytes Received", utils.MetricTypeU64)
	u.noDataMetric = gm.Metric(u.MetricsAt, "No Metric Received", utils.MetricTypeU64)
	u.socketSkipMetric = gm.Metric(u.MetricsAt, "Skip Failed Socket", utils.MetricTypeU64)
	u.socketOpenMetric = gm.Metric(u.MetricsAt, "Socket Opens", utils.MetricTypeU64)
	u.IncorrectRadarMetric = gm.Metric(u.MetricsAt, "Error: Incorrect Radar", utils.MetricTypeU64)
	u.timeMetric = gm.Metric(u.MetricsAt, "Now", utils.MetricTypeU64)
}

func (u *UDPData) Start() {
	u.InitMetrics()
	u.terminate = false
	u.terminated = false
	u.isRunningMetric.SetU32(1, time.Now())

	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)

	u.Connection.Init(u, u.ListenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, u.ReconnectCycle)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, context utils.UDPErrorContext, err error) {
		switch context {
		case utils.UDPErrorOnConnect:
			u.sockOpenFailMetric.AddCount(1, u.Now)
		case utils.UDPErrorOnWriteData:
			u.sockWriteFailMetric.AddCount(1, u.Now)
		case utils.UDPErrorOnReadData:
			u.sockReadFailMetric.AddCount(1, u.Now)
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

	u.isRunningMetric.SetU32(0, time.Now())
	u.terminated = true
}

func (u *UDPData) executeReader() {
	for !u.terminate {
		u.Now = time.Now()
		u.timeMetric.SetTime(u.Now)

		if u.Connection.Listen() {
			u.ClearError()
			u.sockReuseMetric.AddCount(1, u.Now)

			u.BufferLen = u.Connection.ReceiveData(u.Buffer[:], u.Now, u.ReadTimeout)
			if u.BufferLen > 0 {
				u.dataIterationsMetric.AddCount(1, u.Now)
				u.dataBytesMetric.AddCount(uint64(u.BufferLen), u.Now)

				if u.OnData != nil {
					u.OnData(u, u.Connection.FromAddr, u.Buffer[:u.BufferLen])
				}
			} else {
				u.noDataMetric.AddCount(1, u.Now)
			}
		} else {
			u.socketSkipMetric.AddCount(1, u.Now)
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
	u.socketOpenMetric.AddCount(1, u.Now)
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

func (u *UDPData) IsTerminated() bool {
	return u.terminated
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
