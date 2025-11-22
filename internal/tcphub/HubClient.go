package tcphub

import (
	"bufio"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

// HubClient holds a TCP connection with a remote client
// executeWrite sends data from the Radar/via the Hub to the remote client
// executeRead reads data from the remote client(instructions) and dispatches it to the radar
type HubClient struct {
	buffer            PacketBuffer
	connection        net.TCPConn
	writeCache        [2 * utils.Kilobyte]byte
	writeChannel      chan Packet
	doneChannel       chan bool
	host              *HubHost
	stats             HubClientStat
	terminate         bool
	terminateRefCount atomic.Int32
	OnError           func(*HubClient, error)
	OnConnect         func(*HubClient)
	OnDisconnect      func(*HubClient)
}

func (c *HubClient) Init(host *HubHost) {
	c.host = host
	c.terminateRefCount.Store(0)
	c.terminate = true
}

func (c *HubClient) Start(connection net.TCPConn) {
	c.stats.RegisterConnect(connection.RemoteAddr())
	c.buffer.Init(8 * utils.Kilobyte)
	c.connection = connection
	c.terminate = false
	c.terminateRefCount.Store(2)
	c.writeChannel = make(chan Packet, 10)
	c.doneChannel = make(chan bool)

	go c.executeRead()
	go c.executeWrite()

	if c.OnConnect != nil {
		c.OnConnect(c)
	}
}

func (c *HubClient) Stop() {
	if !c.terminate {
		c.doneChannel <- true
	}

	for c.terminateRefCount.Load() > 1 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *HubClient) executeWrite() {
	for {
		select {
		case packet := <-c.writeChannel:
			c.writePacket(packet)

		case <-c.doneChannel:
			c.stats.RegisterDisconnect()
			c.terminate = true
			c.terminateRefCount.Add(-1)
			close(c.writeChannel)
			close(c.doneChannel)
			if c.OnDisconnect != nil {
				c.OnDisconnect(c)
			}
			return
		}
	}
}

func (c *HubClient) Write(packet Packet) bool {
	if c.terminate {
		return false
	}

	if len(c.writeChannel) < cap(c.writeChannel)/2 || packet.Type == PtUdpInstruction {
		c.writeChannel <- packet
		return true
	}
	return false
}

func (c *HubClient) writePacket(packet Packet) {
	if c.terminate {
		return
	}

	slice, err := packet.SaveToBytes(c.writeCache[:])
	if err != nil {
		c.onError(err)
		return
	}

	_, err = c.connection.Write(slice)
	if err != nil {
		c.onError(err)
	}
}

func (c *HubClient) onError(err error) {
	if c.OnError != nil {
		c.OnError(c, err)
	} else {
		log.Err(err).Msg("HubClient::onError")
	}
}

func (c *HubClient) executeRead() {
	var readBuffer [4 * utils.Kilobyte]byte
	reader := bufio.NewReader(&c.connection)

	for !c.terminate {
		bytesRead, err := reader.Read(readBuffer[:])

		if err != nil {
			if err != io.EOF {
				c.onError(err)
				c.doneChannel <- true
				break
			}
		}
		c.stats.RegisterRead(bytesRead)
		c.buffer.PushBytes(readBuffer[:bytesRead])
		var packet Packet

		for c.buffer.Pop(&packet) {
			c.host.dispatcher.writePacket(packet)
		}
	}

	c.terminateRefCount.Add(-1)
}

func (c *HubClient) IsTerminated() bool {
	return c.terminateRefCount.Load() == 0
}
