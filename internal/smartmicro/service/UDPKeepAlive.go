package service

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

const UDPKeepAliveServiceName = "UDP.KeepAlive.Service"

const udpKeepAliveEnabled = "udp.keepalive.enabled"
const udpKeepAliveCallbackIP = "udp.keepalive.callbackip"
const udpKeepAliveCastIP = "udp.keepalive.castip"
const udpKeepAliveCooldown = "udp.keepalive.cooldown"
const udpKeepAliveSendTimeout = "udp.keepalive.send.timeout"
const udpKeepAliveReconnectCycle = "udp.keepalive.reconnect.cycle"
const udpKeepAliveClientID = "udp.keepalive.clientid"
const udpKeepAliveLogRepeatMillis = "udp.keepalive.log.repeat.millis"

// UDPKeepAlive to keep the radar alive
type UDPKeepAlive struct {
	ClientId         uint32    `json:"ClientId"`
	LocalIPAddr      utils.IP4 `json:"LocalIPAddr"`
	MulticastIPAddr  utils.IP4 `json:"MulticastIPAddr"`
	CooldownMs       int       `json:"CooldownMs"`
	ReconnectOnCycle int       `json:"ReconnectOnCycle"`
	SendTimeout      int       `json:"SendTimeout"`
	connection       utils.UDPClientConnection
	buffer           [34]byte
	bufferLen        int
	terminate        bool
	terminated       bool
	now              time.Time
	Metrics          UdpKeepAliveMatrics
	OnTerminate      func(*UDPKeepAlive) `json:"-"`
	utils.ErrorLoggerMixin
}

type UdpKeepAliveMatrics struct {
	IsRunning         *utils.Metric
	DeadlineErr       *utils.Metric
	WriteUDPErr       *utils.Metric
	SendAlive         *utils.Metric
	ConnectUDPErr     *utils.Metric
	ConnectUDPSuccess *utils.Metric
	utils.MetricsInitMixin
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
	s.Metrics.InitMetrics(s.GetServiceName(), &s.Metrics)
	s.terminate = false
	s.terminated = false
	s.connection.OnError = s.onConnectionError
	s.connection.OnConnect = s.onConnectSuccess
	s.connection.Init(s.LocalIPAddr.WithPort(0), s.MulticastIPAddr, s, s.ReconnectOnCycle)

	go s.executeWrite()
}

func (s *UDPKeepAlive) onConnectionError(connection *utils.UDPClientConnection, err error) {
	s.LogError("UDPKeepAlive", err)
	s.Metrics.ConnectUDPErr.Inc(1)
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
	s.Metrics.IsRunning.SetTime()

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

	s.Metrics.IsRunning.SetTime()
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
		s.Metrics.DeadlineErr.Inc(1)
		goto errLabel
	}
	if _, err = cnx.Write(s.buffer[:s.bufferLen]); err != nil {
		s.Metrics.WriteUDPErr.Inc(1)
		goto errLabel
	}

	s.Metrics.SendAlive.Inc(1)
	return

errLabel:
	s.connection.HandleError(err)
	s.connection.Disconnect()
}

func (s *UDPKeepAlive) onConnectSuccess(connection *utils.UDPClientConnection) {
	s.Metrics.ConnectUDPSuccess.Inc(1)
}

func (s *UDPKeepAlive) IsTerminated() bool {
	return s.terminated
}
