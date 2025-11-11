package tcphub

import (
	"net"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

type HubServerDispatcher struct {
	Channel chan Packet
	Addr    utils.IP4
}

func (dispatcher *HubServerDispatcher) execute() {
	for packet := range dispatcher.Channel {
		if packet.Size == 0 {
			break
		}
		dispatcher.handle(packet)
	}

	close(dispatcher.Channel)
}

func (dispatcher *HubServerDispatcher) handle(packet Packet) {
	switch packet.Type {
	case PtUdpInstruction:
		dispatcher.forwardPacket(packet)

	case PtRadarMulticast:
		dispatcher.forwardPacket(packet)
	default:
	}
}

func (dispatcher *HubServerDispatcher) forwardPacket(packet Packet) {
	var conn net.Conn
	var err error

	ip4 := utils.IP4Builder.FromU32(packet.TargetIP4, int(packet.TargetPort))

	remoteAddr := ip4.ToUDPAddr()
	localAddr := dispatcher.Addr.ToUDPAddr()

	if conn, err = net.DialUDP("udp", &localAddr, &remoteAddr); err != nil {
		log.Err(err).
			Str("Addr", ip4.ToString()).
			Msg("Failed to connect to UDP")
		return
	}
	defer conn.Close()

	bytesWritten, err := conn.Write(packet.Data)

	if err != nil {
		log.Err(err).
			Str("Addr", ip4.ToString()).
			Msg("Failed to writePacket to UDP")
	}

	if bytesWritten != len(packet.Data) {
		log.Err(err).
			Str("Addr", ip4.ToString()).
			Msgf("UDP partial writePacket %d of %d bytes", bytesWritten, len(packet.Data))
	}

}

func (dispatcher *HubServerDispatcher) Stop() {
	//dispatcher.Channel <- Packet{}
}
