package server

import (
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/constants"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

// RouterServerService is a wrapper around the RouterServer
// TODO: Put the listAndForward inside the RouterServer
type RouterServerService struct {
	Server     RouterServer
	BindAddr   utils.IP4
	IsEnabled  bool
	MultiAddr  utils.IP4
	Terminate  bool
	Terminated bool
	LastError  error
}

func (r *RouterServerService) InitFromSettings(settings *utils.Settings) {
	r.BindAddr = settings.Basic.GetIP4("router.server.bind.ip.address", utils.IP4Builder.FromString("0.0.0.0:45001"))
	r.MultiAddr = settings.Basic.GetIP4("router.server.multicast.ip.address", utils.IP4Builder.FromString("239.144.0.0:60000"))
}

func (r *RouterServerService) Start(state *utils.State, settings *utils.Settings) {
	state.Set(constants.RouterServerService, r)

	writer := utils.GlobalState.Get(constants.UDPDataServiceName).(*service.UDPDataService)

	if writer != nil {
		writer.RegisterReceiver(r.onReadUDP)
		r.IsEnabled = true
		r.Server.Start(r.BindAddr, r.MultiAddr, writer)

		go r.listenAndForwardMulticast()
	} else {
		r.IsEnabled = false
		log.Warn().Msg("Router Server not started as no UDP Writer found")
	}
}

func (r *RouterServerService) GetServiceName() string {
	return constants.RouterServerService
}

func (r *RouterServerService) listenAndForwardMulticast() {
	multiAddr, err := net.ResolveUDPAddr("udp", r.MultiAddr.String())
	if err != nil {
		log.Error().Err(err).Msg("Error resolving multicast address")
		return
	}

	var conn *net.UDPConn
	var buffer [60]byte

	for !r.Terminate {
		if conn == nil {
			conn, err = net.ListenMulticastUDP("udp", nil, multiAddr)
			if err != nil {
				if !errors.Is(err, r.LastError) {
					log.Error().Err(err).Msg("Error listening multicast UDP")
				}
				time.Sleep(3 * time.Second)
			} else {
				_ = conn.SetReadBuffer(cap(buffer))
			}

		}

		if conn != nil {
			readBytes, readAddr, readErr := conn.ReadFromUDP(buffer[:])

			if readErr != nil {
				log.Error().Err(readErr).Msg("Error reading multicast packet")
				_ = conn.Close()
				conn = nil
			} else {
				remoteIP := utils.IP4Builder.FromAddr(readAddr)

				if remoteIP.Port == 60000 && r.Server.HasConnections() {
					var packetBuffer [256]byte
					var packet tcphub.PacketWrapper
					packet.Init(packetBuffer[:], 0, r.MultiAddr, remoteIP)
					packet.SetPacketType(tcphub.PtRadarMulticast)
					packet.SetData(buffer[:readBytes])
					r.Server.Write(packet.GetPacket())
				}
			}
		}
	}
	r.Terminated = true

	if conn != nil {
		_ = conn.Close()
	}
}

func (r *RouterServerService) onReadUDP(
	dataService *service.UDPDataService,
	addr net.UDPAddr,
	bytes []byte,
) {
	if r.Server.HasConnections() {
		var packetData [2 * utils.Kilobyte]byte

		sourceIP := utils.IP4Builder.FromAddr(&addr)
		packet := tcphub.PacketWrapper{}
		packet.Init(
			packetData[:], 0, dataService.ListenAddr, sourceIP,
		)
		packet.SetData(bytes)
		r.Server.Write(packet.GetPacket())
	}
}
