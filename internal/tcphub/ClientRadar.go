package tcphub

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

const udpBufferSize = 8192

type ClientRadar struct {
	service.MixinDataService
	Client      *Client
	buffer      [udpBufferSize]byte
	bufferLen   int
	radarIPAddr net.UDPAddr
	connection  *net.UDPConn
	sendChannel chan Packet
}

func (r *ClientRadar) Init(client *Client, radarIPAddr utils.IP4) {
	r.LoopGuard = utils.InfiniteLoopGuard{}
	r.RetryGuard = utils.RetryGuard{
		ModCycles: 5,
	}
	r.Client = client
	r.sendChannel = make(chan Packet, 5)
	r.radarIPAddr = radarIPAddr.ToUDPAddr()
	go r.executeReadFromHub()
	go r.executeWriteToUDP()
}

func (r *ClientRadar) SendUDP(packet Packet) {
	if !r.Terminating {
		r.sendChannel <- packet
	}
}

func (r *ClientRadar) SendMulticast(packet Packet) {
	if !r.Terminating {
		r.sendChannel <- packet
	}
}

func (r *ClientRadar) executeWriteToUDP() {
	for packet := range r.sendChannel {
		if packet.Size == 0 {
			break
		}
		if r.connection != nil {
			if packet.Type == PtUdpForward {
				// Writes from the virtual radar address to the client address
				addr := r.Client.ClientAddr.ToUDPAddr()
				bytesWritten, err := r.connection.WriteToUDP(packet.Data, &addr)

				if err != nil {
					log.Err(err).Msgf("Failed to send udp packet to %s", addr.String())
				} else if bytesWritten != len(packet.Data) {
					log.Error().Msgf("Failed to send udp all packet data to %s", addr.String())
				}
			} else if packet.Type == PtRadarMulticast {
				// Send the multicast data
			}
		}
	}
	r.Terminated = true
	r.OnTerminateCallback(r)
	r.closeConnection()
	close(r.sendChannel)
}

func (r *ClientRadar) executeReadFromHub() {
	r.OnStartCallback(r)

	for r.Terminating = false; !r.Terminating; {
		if r.openConnection() {
			r.receiveData()
		}
	}

	r.sendChannel <- Packet{}
}

func (r *ClientRadar) openConnection() bool {
	var err error

	if r.connection != nil {
		return true
	}

	if !r.RetryGuard.ShouldRetry() {
		return false
	}

	if r.connection, err = net.ListenUDP("udp4", &r.radarIPAddr); err != nil {
		goto onErrorLabel
	}

	if err = r.connection.SetReadBuffer(udpBufferSize); err != nil {
		goto onErrorLabel
	}

	r.RetryGuard.Reset()
	r.OnConnectCallback(r)

	return true

onErrorLabel:
	r.OnErrorCallback(r, err)
	r.closeConnection()
	return false
}

func (r *ClientRadar) closeConnection() {
	if r.connection != nil {
		r.OnDisconnectCallback(r)
		_ = r.connection.Close()
		r.connection = nil
	}
}

func (r *ClientRadar) receiveData() {
	var err error
	var fromAddr *net.UDPAddr
	var packet Packet

	deadline := time.Now().Add(time.Duration(1) * time.Second)

	if err = r.connection.SetReadDeadline(deadline); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return
		}
		goto onErrorLabel
	}
	if r.bufferLen, fromAddr, err = r.connection.ReadFromUDP(r.buffer[:]); err != nil {
		goto onErrorLabel
	}

	// Udp is complete
	packet = NewPacket(r.buffer[:r.bufferLen])
	packet.SetTarget(fromAddr)
	r.Client.SendToServer(packet)

	return

onErrorLabel:
	r.OnErrorCallback(r, err)
	r.closeConnection()
}
