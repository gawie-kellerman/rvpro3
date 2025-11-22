package tcphub

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

const udpBufferSize = 8 * utils.Kilobyte

// VirtualRadar hosts a UDP Server on the Client Desktop that serves
// to intercept instructions sent from the desktop to the radar
// as well as relaying UDP from the actual radar to the desktop
type VirtualRadar struct {
	host      *VirtualHost
	buffer    [udpBufferSize]byte
	bufferLen int

	radarAddress      utils.IP4
	writeChannel      chan Packet
	doneChannel       chan bool
	terminate         bool
	terminateRefCount atomic.Int32
	udpServer         utils.UDPServerConnection
	OnReadError       func(*VirtualRadar, error)
	OnWriteError      func(*VirtualRadar, error)
}

func (s *VirtualRadar) Start(host *VirtualHost, listenAddr utils.IP4) {
	s.radarAddress = listenAddr
	s.host = host
	s.terminate = false
	s.terminateRefCount.Store(2)
	s.writeChannel = make(chan Packet, 10)
	s.doneChannel = make(chan bool, 1)
	s.udpServer.Init(s, listenAddr, 32*utils.Kilobyte, 32*utils.Kilobyte, 3)

	go s.executeRead()
	go s.executeWrite()
}

func (s *VirtualRadar) Stop() {
	if !s.terminate {
		s.doneChannel <- true
	}

	for s.terminateRefCount.Load() > 1 {
		time.Sleep(1 * time.Second)
	}
}

func (s *VirtualRadar) executeRead() {
	var err error
	for !s.terminate {

		if s.udpServer.Listen() {
			if cnx := s.udpServer.GetConnection(); cnx != nil {
				if s.bufferLen, err = cnx.Read(s.buffer[:]); err != nil {
					if s.OnReadError != nil {
						s.OnReadError(s, err)
					}
					s.udpServer.Close()
				} else {
					packet := NewPacket(s.buffer[:s.bufferLen])
					packet.SetTarget(s.udpServer.GetConnection().LocalAddr().(*net.UDPAddr))
					s.host.WriteToHub(packet)
				}
			}
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	s.terminateRefCount.Add(-1)
}

func (s *VirtualRadar) executeWrite() {
	for {
		select {
		case packet := <-s.writeChannel:
			s.writePacket(packet)

		case <-s.doneChannel:
			s.terminate = true
			s.terminateRefCount.Add(-1)
			close(s.writeChannel)
			close(s.doneChannel)
			return
		}
	}

}

// WriteToDesktop  writes the packet to the desktop ip (192.168.11.2)
// from the radar ip (e.g. 192.168.11.12..15)
func (s *VirtualRadar) WriteToDesktop(packet Packet) {
	if !s.terminate {
		s.writeChannel <- packet
	} else {
		fmt.Println("talk to the hand")
	}
}

// writePacket  writes the packet to the desktop ip (192.168.11.2)
// from the radar ip (e.g. 192.168.11.12..15)
func (s *VirtualRadar) writePacket(packet Packet) {
	var err error
	var buffer [udpBufferSize]byte
	var slice []byte

	cnx := s.udpServer.GetConnection()
	if cnx != nil {
		if slice, err = packet.SaveToBytes(buffer[:]); err != nil {
			goto handleError
		}
		if _, err = cnx.WriteToUDP(slice, &s.host.desktopIPUDP); err != nil {
			goto handleError
		}
	}

	return

handleError:
	s.handleWriteError(err)
}

func (s *VirtualRadar) handleWriteError(err error) {
	if s.OnWriteError != nil {
		s.OnWriteError(s, err)
	} else {
		log.Err(err).Msgf("writer error on virtual radar %v", s.radarAddress)
	}
}
