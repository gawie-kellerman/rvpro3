package tcphub

import (
	"net"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

// HubDispatcher sends UDP data to the attached radars
// The VirtualRadar.executeRead receives a message (e.g. on 192.168.11.12:55555) and
// sends it to VirtualHost, which sends it to HubClient.  HubClient sends it
// to HubHost which dispenses it to HubDispatcher
type HubDispatcher struct {
	stats        HubDispatcherStat
	writeChannel chan Packet
	doneChannel  chan bool
	terminate    bool
	LocalAddr    utils.IP4
	OnError      func(*HubDispatcher, error)
}

func (hd *HubDispatcher) Start(localAddr utils.IP4) {
	hd.LocalAddr = localAddr
	hd.writeChannel = make(chan Packet, 10)
	hd.doneChannel = make(chan bool, 1)
	go hd.executeWrite()
}

func (hd *HubDispatcher) Stop() {
	hd.doneChannel <- true
}

func (hd *HubDispatcher) executeWrite() {
	for {
		select {
		case packet := <-hd.writeChannel:
			hd.writePacket(packet)
		case <-hd.doneChannel:
			hd.terminate = true
			close(hd.writeChannel)
			close(hd.doneChannel)
			return
		}
	}
}

func (hd *HubDispatcher) writePacket(packet Packet) {
	var err error
	var conn *net.UDPConn
	packetLen := uint32(len(packet.Data))

	localAddr := hd.LocalAddr.ToUDPAddr()
	remoteAddr := packet.GetTargetIP().ToUDPAddr()

	if conn, err = net.DialUDP("udp4", &localAddr, &remoteAddr); err != nil {
		hd.stats.RegisterError(packetLen)
		hd.onError(err)
		return
	} else {
		defer conn.Close()

		if _, err = conn.Write(packet.Data); err != nil {
			hd.stats.RegisterError(packetLen)
			hd.onError(err)
			return
		} else {
			hd.stats.RegisterWrite(packetLen)
		}
	}
}

func (hd *HubDispatcher) onError(err error) {
	if hd.OnError != nil {
		hd.OnError(hd, err)
	} else {
		log.Err(err).Msg("HubDispatcher error")
	}
}
