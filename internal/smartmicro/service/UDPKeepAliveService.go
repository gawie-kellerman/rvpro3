package service

import (
	"time"

	"rvpro3/radarvision.com/internal/general"
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

// UDPKeepAliveService to keep the radar alive
type UDPKeepAliveService struct {
	Now              time.Time `json:"Now"`
	IsEnabled        bool
	ClientId         uint32             `json:"ClientId"`
	LocalIPAddr      utils.IP4          `json:"LocalIPAddr"`
	MulticastIPAddr  utils.IP4          `json:"MulticastIPAddr"`
	CooldownDuration utils.Milliseconds `json:"CooldownDuration"`
	SendTimeout      utils.Milliseconds `json:"SendTimeout"`
	ReconnectOnCycle int                `json:"ReconnectOnCycle"`
	CurrentErr       utils.ErrorLoggerMixin
	Metrics          UdpKeepAliveMatrics
	OnTerminate      func(*UDPKeepAliveService) `json:"-"`
	connection       utils.UDPClientConnection
	buffer           [34]byte
	bufferLen        int
	terminate        bool
	terminated       bool
}

type UdpKeepAliveMatrics struct {
	SendAliveCount       *utils.Metric
	SocketConnectSuccess *utils.Metric
	ErrorSocketDeadline  *utils.Metric
	ErrorSocketWrite     *utils.Metric
	ErrorSocketConnect   *utils.Metric
	utils.MetricsInitMixin
}

func (s *UDPKeepAliveService) InitFromSettings(settings *utils.Settings) {
	s.IsEnabled = settings.Basic.GetBool(udpKeepAliveEnabled, true)
	s.ClientId = uint32(settings.Basic.GetInt(udpKeepAliveClientID, 0x1000001))
	s.CurrentErr.RepeatDuration = settings.Basic.GetMilliseconds(udpKeepAliveLogRepeatMillis, 60000)
	s.SendTimeout = settings.Basic.GetMilliseconds(udpKeepAliveSendTimeout, 1000)
	s.ReconnectOnCycle = settings.Basic.GetInt(udpKeepAliveReconnectCycle, 5)
	s.CooldownDuration = settings.Basic.GetMilliseconds(udpKeepAliveCooldown, 1000)
	s.LocalIPAddr = settings.Basic.GetIP4(udpKeepAliveCallbackIP, utils.IP4Builder.FromString("192.168.11.2:55555"))
	s.MulticastIPAddr = settings.Basic.GetIP4(udpKeepAliveCastIP, utils.IP4Builder.FromString("239.144.0.0:60000"))
}

func (s *UDPKeepAliveService) Start(state *utils.State, config *utils.Settings) {
	if !general.ServiceHelper.ShouldStart(state, config, s) {
		return
	}

	if !s.IsEnabled {
		return
	}
	s.init()
	go s.executeWrite()
}

func (s *UDPKeepAliveService) GetServiceName() string {
	return UDPKeepAliveServiceName
}

func (s *UDPKeepAliveService) GetServiceNames() []string {
	return nil
}

func (s *UDPKeepAliveService) init() {
	s.Metrics.InitMetrics(s.GetServiceName(), &s.Metrics)
	s.terminate = false
	s.terminated = false
	s.connection.OnError = s.onConnectionError
	s.connection.OnConnect = s.onConnectSuccess
	s.connection.Init(s.LocalIPAddr.WithPort(0), s.MulticastIPAddr, s, s.ReconnectOnCycle)
}

func (s *UDPKeepAliveService) onConnectionError(_ *utils.UDPClientConnection, err error) {
	s.CurrentErr.LogErrorAt(time.Now(), s.GetServiceName(), err)
	s.Metrics.ErrorSocketConnect.Inc(1)
}

func (s *UDPKeepAliveService) Stop() {
	s.terminate = true
	for !s.terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *UDPKeepAliveService) Run() {
	go s.executeWrite()
}

func (s *UDPKeepAliveService) executeWrite() {
	s.initBuffer()

	for !s.terminate {
		s.Now = time.Now()

		if s.connection.Connect() {
			s.sendAlive()
		}

		if !s.terminate {
			s.CooldownDuration.Sleep()
		}
	}
	s.connection.Disconnect()
	if s.OnTerminate != nil {
		s.OnTerminate(s)
	}

	s.terminated = true
}

func (s *UDPKeepAliveService) initBuffer() {
	alive := port.NewClientKeepAlive(
		s.ClientId,
		s.LocalIPAddr.ToU32(),
		uint16(s.LocalIPAddr.Port),
	)

	writer := utils.NewFixedBuffer(s.buffer[:], 0, 0)
	alive.Write(&writer)
	s.bufferLen = writer.WritePos
}

func (s *UDPKeepAliveService) sendAlive() {
	cnx := s.connection.GetConnection()

	if cnx == nil {
		return
	}

	var err error
	timeout := s.Now.Add(time.Duration(s.SendTimeout))
	err = cnx.SetWriteDeadline(timeout)
	if err != nil {
		s.Metrics.ErrorSocketDeadline.Inc(1)
		goto errLabel
	}
	if _, err = cnx.Write(s.buffer[:s.bufferLen]); err != nil {
		s.Metrics.ErrorSocketWrite.Inc(1)
		goto errLabel
	}

	s.Metrics.SendAliveCount.Inc(1)
	return

errLabel:
	s.connection.HandleError(err)
	s.connection.Disconnect()
}

func (s *UDPKeepAliveService) onConnectSuccess(connection *utils.UDPClientConnection) {
	s.CurrentErr.Clear()
	s.Metrics.SocketConnectSuccess.Inc(1)
}

func (s *UDPKeepAliveService) IsTerminated() bool {
	return s.terminated
}
