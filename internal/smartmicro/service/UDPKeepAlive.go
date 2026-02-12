package service

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

const UDPKeepAliveServiceName = "UDP.KeepAlive.Service"

const udpKeepAliveEnabled = "UDP.KeepAlive.Enabled"
const udpKeepAliveCallbackIP = "UDP.KeepAlive.CallbackIP"
const udpKeepAliveCastIP = "UDP.KeepAlive.CastIP"
const udpKeepAliveCooldown = "UDP.KeepAlive.Cooldown"
const udpKeepAliveSendTimeout = "UDP.KeepAlive.Timeout"
const udpKeepAliveReconnectCycle = "UDP.KeepAlive.ReconnectCycle"
const udpKeepAliveClientID = "UDP.KeepAlive.ClientID"
const udpKeepAliveLogRepeatMillis = "UDP.KeepAlive.RepeatMillis"

// UDPKeepAlive to keep the radar alive
type UDPKeepAlive struct {
	ClientId         uint32
	LocalIPAddr      utils.IP4
	MulticastIPAddr  utils.IP4
	CooldownMs       int
	ReconnectOnCycle int
	SendTimeout      int
	connection       utils.UDPClientConnection
	buffer           [34]byte
	bufferLen        int
	terminate        bool
	terminated       bool
	now              time.Time
	Metrics          udpKeepAliveMatrics
	OnTerminate      func(*UDPKeepAlive)
	utils.ErrorLoggerMixin
}

type udpKeepAliveMatrics struct {
	IsRunningMetric         *utils.Metric
	DeadlineErrMetric       *utils.Metric
	WriteUDPErrMetric       *utils.Metric
	SendAliveMetric         *utils.Metric
	ConnectUDPErrMetric     *utils.Metric
	ConnectUDPSuccessMetric *utils.Metric
}

func (s *udpKeepAliveMatrics) InitMetrics(serviceName string) {
	gm := &utils.GlobalMetrics
	sn := serviceName
	s.IsRunningMetric = gm.Metric(sn, "Is Running", utils.MetricTypeU32)
	s.ConnectUDPErrMetric = gm.Metric(sn, "Error: UDP Connect", utils.MetricTypeU32)
	s.DeadlineErrMetric = gm.U64(sn, "Error: SetRaw UDP Deadline")
	s.WriteUDPErrMetric = gm.U64(sn, "Error: UDP WritePacket")
	s.SendAliveMetric = gm.U64(sn, "Send Alive")
	s.ConnectUDPSuccessMetric = gm.U64(sn, "UDP Connect Success")
}

func (s *UDPKeepAlive) SetupDefaults(config *utils.Settings) {
	config.SetSettingAsBool(udpKeepAliveEnabled, true)
	config.SetSettingAsInt(udpKeepAliveClientID, 0x1000001)
	config.SetSettingAsInt(udpKeepAliveLogRepeatMillis, 60000)
	config.SetSettingAsInt(udpKeepAliveSendTimeout, 1000)
	config.SetSettingAsInt(udpKeepAliveReconnectCycle, 5)
	config.SetSettingAsInt(udpKeepAliveCooldown, 1000)
	config.SetSettingAsStr(udpKeepAliveCallbackIP, "192.168.11.2:55555")
	config.SetSettingAsStr(udpKeepAliveCastIP, "239.144.0.0:60000")
}

func (s *UDPKeepAlive) SetupAndStart(state *utils.State, config *utils.Settings) {
	if !config.GetSettingAsBool(udpKeepAliveEnabled) {
		return
	}

	s.InitFromSettings(config)
	s.Start()
	state.Set(s.GetServiceName(), s)
}

func (s *UDPKeepAlive) InitFromSettings(config *utils.Settings) {
	s.LocalIPAddr = config.GetSettingAsIP(udpKeepAliveCallbackIP)
	s.MulticastIPAddr = config.GetSettingAsIP(udpKeepAliveCastIP)
	s.CooldownMs = config.GetSettingAsInt(udpKeepAliveCooldown)
	s.ReconnectOnCycle = config.GetSettingAsInt(udpKeepAliveReconnectCycle)
	s.SendTimeout = config.GetSettingAsInt(udpKeepAliveSendTimeout)
	s.ClientId = uint32(config.GetSettingAsInt(udpKeepAliveClientID))
	s.LogRepeatMillis = config.GetSettingAsMillis(udpKeepAliveLogRepeatMillis)
}

func (s *UDPKeepAlive) GetServiceName() string {
	return UDPKeepAliveServiceName
}

func (s *UDPKeepAlive) GetServiceNames() []string {
	return nil
}

func (s *UDPKeepAlive) Start() {
	s.Metrics.InitMetrics(s.GetServiceName())
	s.terminate = false
	s.terminated = false
	s.connection.OnError = s.onConnectionError
	s.connection.OnConnect = s.onConnectSuccess
	s.connection.Init(s.LocalIPAddr.WithPort(0), s.MulticastIPAddr, s, s.ReconnectOnCycle)

	go s.executeWrite()
}

func (s *UDPKeepAlive) onConnectionError(connection *utils.UDPClientConnection, err error) {
	s.LogError("UDPKeepAlive", err)
	s.Metrics.ConnectUDPErrMetric.Inc(s.now)
}

func (s *UDPKeepAlive) Stop() {
	s.terminate = true
	for !s.terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *UDPKeepAlive) Run() {
	go s.executeWrite()
}

func (s *UDPKeepAlive) executeWrite() {
	s.Metrics.IsRunningMetric.SetU32(1, time.Now())
	s.initBuffer()

	for !s.terminate {
		s.now = time.Now()

		if s.connection.Connect() {
			s.sendAlive()
		}

		if !s.terminate {
			time.Sleep(time.Duration(s.CooldownMs) * time.Millisecond)
		}
	}
	s.connection.Disconnect()
	if s.OnTerminate != nil {
		s.OnTerminate(s)
	}

	s.Metrics.IsRunningMetric.SetU32(0, time.Now())
	s.terminated = true
}

func (s *UDPKeepAlive) initBuffer() {
	alive := port.NewClientKeepAlive(
		s.ClientId,
		s.LocalIPAddr.ToU32(),
		uint16(s.LocalIPAddr.Port),
	)

	writer := utils.NewFixedBuffer(s.buffer[:], 0, 0)
	alive.Write(&writer)
	s.bufferLen = writer.WritePos
}

func (s *UDPKeepAlive) sendAlive() {
	cnx := s.connection.GetConnection()

	if cnx == nil {
		return
	}

	var err error
	timeout := s.now.Add(time.Duration(s.SendTimeout) * time.Millisecond)
	err = cnx.SetWriteDeadline(timeout)
	if err != nil {
		s.Metrics.DeadlineErrMetric.Inc(s.now)
		goto errLabel
	}
	if _, err = cnx.Write(s.buffer[:s.bufferLen]); err != nil {
		s.Metrics.WriteUDPErrMetric.Inc(s.now)
		goto errLabel
	}

	s.Metrics.SendAliveMetric.Inc(s.now)
	return

errLabel:
	s.connection.HandleError(err)
	s.connection.Disconnect()
}

func (s *UDPKeepAlive) onConnectSuccess(connection *utils.UDPClientConnection) {
	s.Metrics.ConnectUDPSuccessMetric.Inc(s.now)
}

func (s *UDPKeepAlive) IsTerminated() bool {
	return s.terminated
}
