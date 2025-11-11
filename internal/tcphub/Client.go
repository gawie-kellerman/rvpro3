package tcphub

import (
	"errors"
	"io"
	"net"
	"os"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

var GlobalTCPHubClient Client

type Client struct {
	service.MixinDataService
	Buffer        PacketBuffer
	ServerAddr    utils.IP4
	ClientRadar   [4]ClientRadar
	connection    *net.TCPConn
	readBuffer    [16 * utils.Kilobyte]byte
	readBufferLen int
	writeChannel  chan Packet
	ClientAddr    utils.IP4
}

func (c *Client) Start(serverAddr utils.IP4, clientAddr utils.IP4) {
	c.Init(serverAddr, clientAddr)

	go c.executeWriteToHub()
	go c.executeReadFromHub()
}

func (c *Client) Init(serverAddr utils.IP4, clientAddr utils.IP4) {
	c.ServerAddr = serverAddr
	c.ClientAddr = clientAddr
	c.writeChannel = make(chan Packet)
	c.Buffer.Init(16 * utils.Kilobyte)

	for i := 0; i < len(c.ClientRadar); i++ {
		radar := &c.ClientRadar[i]
		radar.Init(c, utils.RadarIPOf(i))
	}
}

// executeReadFromHub reads data from the port as sent by the
// tcphub.Server
func (c *Client) executeReadFromHub() {
	c.OnStartCallback(c)

	for c.Terminating = false; !c.Terminating; {
		if c.openConnection() {
			c.readData()
		}
	}

	c.closeConnection()
	c.OnTerminateCallback(c)
	c.writeChannel <- Packet{}
	c.Terminated = true
}

func (c *Client) executeWriteToHub() {
	for packet := range c.writeChannel {
		if packet.Size == 0 {
			break
		}
		if c.openConnection() {
			c.sendData(packet)
		}
	}
}

func (c *Client) closeConnection() {
	if c.connection != nil {
		c.OnDisconnectCallback(c)
		_ = c.connection.Close()
		c.connection = nil
	}
}

func (c *Client) openConnection() bool {
	var err error
	var serverAddr net.TCPAddr

	if c.connection != nil {
		return true
	}

	if !c.RetryGuard.ShouldRetry() {
		return false
	}

	//if serverAddr, err = net.ResolveTCPAddr("tcp4", c.ServerAddr); err != nil {
	//	goto onErrorLabel
	//}
	serverAddr = c.ServerAddr.ToTCPAddr()
	if c.connection, err = net.DialTCP("tcp4", nil, &serverAddr); err != nil {
		goto onErrorLabel
	}

	c.RetryGuard.Reset()
	c.OnConnectCallback(c)
	return true

onErrorLabel:
	c.OnErrorCallback(c, err)
	c.closeConnection()
	return false
}

func (c *Client) sendData(packet Packet) {
	var buffer [4 * utils.Kilobyte]byte
	slice, err := packet.SaveToBytes(buffer[:])
	if err != nil {
		c.OnErrorCallback(c, err)
		return
	}

	if _, err = c.connection.Write(slice); err != nil {
		c.OnErrorCallback(c, err)
	}
}

func (c *Client) readData() {
	var err error

	deadline := time.Now().Add(3 * time.Second)
	if err = c.connection.SetReadDeadline(deadline); err != nil {
		goto onErrorLabel
	}

	if c.readBufferLen, err = c.connection.Read(c.readBuffer[:]); err != nil {
		goto onErrorLabel
	}

	if c.readBufferLen > 0 {
		c.Buffer.PushBytes(c.readBuffer[:c.readBufferLen])
		var packet Packet
		for c.Buffer.Pop(&packet) {
			switch packet.Type {
			case PtUdpForward:
				c.doUdpForwardReceived(packet)

			case PtStats:
				c.doStatsReceived(packet)

			case PtRadarMulticast:
				c.doRadarMulticastReceived(packet)

			case PtUnknown:
				c.doUnknown(packet)

			case PtUdpInstruction:
				c.doUdpInstruction(packet)
			}
		}
	}

	return

onErrorLabel:
	if err != io.EOF && !errors.Is(err, os.ErrDeadlineExceeded) {
		c.OnErrorCallback(c, err)
		c.closeConnection()
	}
}

func (c *Client) doUdpForwardReceived(packet Packet) {
	index := utils.RadarIndexOf(packet.TargetIP4)

	if index == -1 {
		c.OnErrorCallback(c, errors.New("invalid packet received"))
		return
	}

	c.ClientRadar[index].SendUDP(packet)
}

func (c *Client) doStatsReceived(packet Packet) {

}

func (c *Client) doRadarMulticastReceived(packet Packet) {
	index := utils.RadarIndexOf(packet.TargetIP4)

	if index == -1 {
		c.OnErrorCallback(c, errors.New("invalid packet received"))
		return
	}

	c.ClientRadar[index].SendMulticast(packet)
}

func (c *Client) doUdpInstruction(packet Packet) {
	// Should never get here, as the instruction recognition is
	// from Client to Server only.  From Server to Client it
	// is simply a UdpForward
}

func (c *Client) doUnknown(packet Packet) {

}

func (c *Client) SendToServer(packet Packet) {
	// Send back to the server
	if c.Terminating {
		return
	}
	c.writeChannel <- packet
}

func (c *Client) Stop() {
	c.Terminating = true

	for i := 0; i < len(c.ClientRadar); i++ {
		radar := &c.ClientRadar[i]
		if !radar.Terminating {
			radar.Terminating = true
		}
	}

}
