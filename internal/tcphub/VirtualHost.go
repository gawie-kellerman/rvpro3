package tcphub

import (
	"errors"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

var GlobalVirtualHost VirtualHost

type VirtualHost struct {
	OnError           func(*VirtualHost, error)
	buffer            PacketBuffer
	radars            [4]VirtualRadar
	tcpConn           utils.TCPClientConnection
	readBuffer        [16 * utils.Kilobyte]byte
	desktopIP         utils.IP4
	desktopIPUDP      net.UDPAddr
	readBufferLen     int
	writeToHubChannel chan Packet
	doneChannel       chan bool
	terminate         bool
	terminateRefCount atomic.Int32
}

const maxContinuousDeadlines = 10

var ErrContinuousDeadlines = errors.New("continuous deadlines reached")
var ErrInvalidRadar = errors.New("packet from invalid radar received")

func (vh *VirtualHost) Start(hubHostAddress utils.IP4) {
	vh.desktopIP = utils.IP4Builder.FromString("192.168.11.1:55555")
	vh.desktopIPUDP = vh.desktopIP.ToUDPAddr()
	vh.terminate = false
	vh.terminateRefCount.Store(2)
	vh.writeToHubChannel = make(chan Packet, 10)
	vh.doneChannel = make(chan bool, 1)
	vh.buffer = NewBuffer(16 * utils.Kilobyte)
	vh.tcpConn.Init(vh, hubHostAddress, 5, 8*utils.Kilobyte, 4*utils.Kilobyte)

	for i := range vh.radars {
		radarIP := utils.RadarIPOf(i)
		radar := &vh.radars[i]
		radar.Start(vh, radarIP)
	}

	go vh.executeReadFromHub()
	go vh.executeWriteToHub()
}

func (vh *VirtualHost) Stop() {
	if !vh.terminate {
		vh.doneChannel <- true
	}

	for vh.terminateRefCount.Load() > 1 {
		time.Sleep(1 * time.Second)
	}
}

func (vh *VirtualHost) executeReadFromHub() {
	var err error
	contDeadlines := 0

	for !vh.terminate {
		if vh.tcpConn.Connect() {
			if cnx := vh.tcpConn.GetConnection(); cnx != nil {
				_ = cnx.SetReadDeadline(time.Now().Add(1 * time.Second))

				if vh.readBufferLen, err = cnx.Read(vh.readBuffer[:]); err != nil {
					if errors.Is(err, os.ErrDeadlineExceeded) {
						contDeadlines++
						if contDeadlines > maxContinuousDeadlines {
							vh.onError(ErrContinuousDeadlines)
							vh.tcpConn.Disconnect()
						}
					} else {
						contDeadlines = 0
						vh.onError(err)
						vh.tcpConn.Disconnect()
					}
				} else {
					// data successfully read
					contDeadlines = 0
					vh.buffer.PushBytes(vh.readBuffer[:vh.readBufferLen])

					var packet Packet
					for vh.buffer.Pop(&packet) {
						switch packet.Type {
						case PtUdpForward:
							vh.doUdpForwardReceived(packet)

						case PtStats:
							vh.doStatsReceived(packet)

						case PtRadarMulticast:
							vh.doRadarMulticastReceived(packet)

						case PtUdpInstruction:
							vh.doUdpInstructionReceived(packet)

						case PtServerClosesConnection:
							vh.doServerClosesConnectionReceived(packet)

						default:
							vh.doUnknownReceived(packet)
						}
					}
				}
			}
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	vh.terminateRefCount.Add(-1)
}

func (vh *VirtualHost) executeWriteToHub() {
	for {
		select {
		case packet := <-vh.writeToHubChannel:
			vh.writePacketToHub(packet)

		case <-vh.doneChannel:
			vh.terminate = true
			vh.terminateRefCount.Add(-1)

			for i := range vh.radars {
				radar := &vh.radars[i]
				radar.Stop()
			}

			close(vh.writeToHubChannel)
			close(vh.doneChannel)
		}
	}
}

func (vh *VirtualHost) writePacketToHub(packet Packet) {
	if cnx := vh.tcpConn.GetConnection(); cnx != nil {
		var packetBytes [4 * utils.Kilobyte]byte
		if slice, err := packet.SaveToBytes(packetBytes[:]); err != nil {
			vh.onError(err)
		} else {
			if _, err = cnx.Write(slice); err != nil {
				vh.onError(err)
			}
		}
	}
}

func (vh *VirtualHost) WriteToHub(packet Packet) {
	if !vh.terminate {
		vh.writeToHubChannel <- packet
	}
}

func (vh *VirtualHost) onError(err error) {
	if vh.OnError != nil {
		vh.OnError(vh, err)
	} else {
		log.Err(err).Msg("virtual host error")
	}
}

func (vh *VirtualHost) doUdpForwardReceived(packet Packet) {
	index := utils.RadarIndexOf(packet.TargetIP4)

	if index == -1 {
		vh.onError(ErrInvalidRadar)
		return
	}

	vh.radars[index].WriteToDesktop(packet)
}

func (vh *VirtualHost) doUdpInstructionReceived(packet Packet) {
	vh.doUdpForwardReceived(packet)
}

func (vh *VirtualHost) doUnknownReceived(packet Packet) {

}

func (vh *VirtualHost) doServerClosesConnectionReceived(packet Packet) {

}

func (vh *VirtualHost) doRadarMulticastReceived(packet Packet) {

}

func (vh *VirtualHost) doStatsReceived(packet Packet) {

}
