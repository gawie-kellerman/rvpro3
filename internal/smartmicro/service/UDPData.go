package service

import (
	"net"
	"sync/atomic"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/instrumentation"
	"rvpro3/radarvision.com/utils"
)

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
	timeMetric           *instrumentation.Metric
	connectErrorMetric   *instrumentation.Metric
	writeFailMetric      *instrumentation.Metric
	socketReadFailMetric *instrumentation.Metric
	socketUseMetric      *instrumentation.Metric
	dataIterationsMetric *instrumentation.Metric
	dataBytesMetric      *instrumentation.Metric
	noDataMetric         *instrumentation.Metric
	socketSkipMetric     *instrumentation.Metric
	socketOpenMetric     *instrumentation.Metric
	utils.ErrorLoggerMixin
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

func (u *UDPData) metric(m instrumentation.UDPMetric) *instrumentation.Metric {
	return instrumentation.GlobalUDPMetrics.GetRel(int(m))
}

func (u *UDPData) InitMetrics() {
	u.MetricsAt = "UDP.Data"
	gm := &instrumentation.GlobalMetrics
	u.connectErrorMetric = gm.Metric(u.MetricsAt, "Connect Error", instrumentation.MetricTypeU64)
	u.writeFailMetric = gm.Metric(u.MetricsAt, "Write Fail", instrumentation.MetricTypeU64)
	u.socketReadFailMetric = gm.Metric(u.MetricsAt, "Socket Read Fail", instrumentation.MetricTypeU64)
	u.socketUseMetric = gm.Metric(u.MetricsAt, "Socket Use", instrumentation.MetricTypeU64)
	u.dataIterationsMetric = gm.Metric(u.MetricsAt, "Data Iterations", instrumentation.MetricTypeU64)
	u.dataBytesMetric = gm.Metric(u.MetricsAt, "Data Bytes", instrumentation.MetricTypeU64)
	u.noDataMetric = gm.Metric(u.MetricsAt, "No Data", instrumentation.MetricTypeU64)
	u.socketSkipMetric = gm.Metric(u.MetricsAt, "Socket Skip", instrumentation.MetricTypeU64)
	u.socketOpenMetric = gm.Metric(u.MetricsAt, "Socket Open", instrumentation.MetricTypeU64)
	u.timeMetric = gm.Metric(u.MetricsAt, "Time", instrumentation.MetricTypeU64)
}

func (u *UDPData) Start() {
	u.InitMetrics()

	u.TerminateRefCount.Store(2)
	u.doneChannel = make(chan bool)
	u.writeChannel = make(chan UDPSendData, 4)

	u.Connection.Init(u, u.ListenAddr, 4*utils.Kilobyte, 4*utils.Kilobyte, u.ReconnectCycle)

	u.Connection.OnError = func(connection *utils.UDPServerConnection, context utils.UDPErrorContext, err error) {
		switch context {
		case utils.UDPErrorOnConnect:
			u.connectErrorMetric.AddCount(1, u.Now)
		case utils.UDPErrorOnWriteData:
			u.writeFailMetric.AddCount(1, u.Now)
		case utils.UDPErrorOnReadData:
			u.socketReadFailMetric.AddCount(1, u.Now)
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
}

func (u *UDPData) executeReader() {
	for !u.terminate {
		u.Now = time.Now()
		u.timeMetric.SetTime(u.Now)

		if u.Connection.Listen() {
			u.ClearError()
			u.socketUseMetric.AddCount(1, u.Now)

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

type UDPSendData struct {
	Address utils.IP4
	Data    []byte
}
