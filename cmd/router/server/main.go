package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/router/server"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

var config utils.Settings

func main() {
	config.Init()

	hubIP := utils.IP4Builder.FromString("0.0.0.0:45001")
	obj := mainType{}
	obj.Start(hubIP)

	// Run for 10 minutes
	fmt.Println("Press Ctrl-C to quit.")
	time.Sleep(600 * time.Minute)
}

type mainType struct {
	HubIP     utils.IP4
	Wg        sync.WaitGroup
	Server    server.Server
	KeepAlive service.UDPKeepAlive
	Data      service.UDPData
}

func (m *mainType) WritePacket(packetData []byte) error {
	packet := tcphub.PacketWrapper{
		Buffer: packetData[:],
	}

	addr := packet.GetTargetIP4().ToUDPAddr()

	return m.Data.Connection.WriteData(addr, packet.GetData())
}

func (m *mainType) Start(hubIP utils.IP4) {
	m.KeepAlive.SetupDefaults(&config)
	m.KeepAlive.InitFromSettings(&config)

	m.Data.SetupDefaults(&config)
	m.Data.InitFromConfig(&config)

	m.HubIP = hubIP
	m.Data.OnData = m.onRadarDataReceived
	m.KeepAlive.Start()
	m.Data.Start()
	m.Server.Start(m.HubIP, m)

	go m.runMulticast(m.KeepAlive.LocalIPAddr)
}

func (m *mainType) onRadarDataReceived(service *service.UDPData, addr net.UDPAddr, bytes []byte) {
	sourceIP := utils.IP4Builder.FromAddr(&addr)
	var packetData [2 * utils.Kilobyte]byte
	packet := tcphub.PacketWrapper{}
	packet.Init(packetData[:], 0, service.ListenAddr, sourceIP)
	packet.SetData(bytes)
	//packet.Dump("Downward")
	m.Server.Write(packet.GetPacket())
}

func (m *mainType) runMulticast(targetIP utils.IP4) {
	sourceIP := utils.IP4Builder.FromString("239.144.0.0:60000")
	addr, err := net.ResolveUDPAddr("udp", sourceIP.String())
	if err != nil {
		log.Err(err).Msg("Failed to resolve UDP address")
		return
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Err(err).Msg("Failed to listen multicast UDP")
		return
	}
	defer conn.Close()

	_ = conn.SetReadBuffer(4096)

	for {
		buffer := make([]byte, 4096)
		noBytes, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Err(err).Msg("Failed to read from UDP")
			return
		}

		remoteIP := utils.IP4Builder.FromAddr(remoteAddr)
		var packetBuffer [2000]byte
		packet := tcphub.PacketWrapper{}
		if remoteIP.Port == 60000 {
			fmt.Println("Broadcast Remote", remoteIP, "Source", sourceIP)
			packet.Init(packetBuffer[:], 0, sourceIP, remoteIP)
			packet.SetPacketType(tcphub.PtRadarMulticast)
			packet.SetData(buffer[:noBytes])
			m.Server.Write(packet.GetPacket())
		}
		//log.Info().Msgf("Read %d bytes from %s", noBytes, srcAddr)
	}
}
