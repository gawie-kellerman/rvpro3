package client

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

const hubClient = "Hub.Client"

// Client connects to HubServer.  It takes HubServer TCP traffic, interprets it
// and forwards it to the local machine mimicking the UDP data like if it was a radar
// by hosting a UDP server at the address.
// TODO: Propagate Radar Multicast as Multicast via the Client!
type Client struct {
	RemoteAddr   utils.IP4
	packetQueue  utils.QueueBuffer
	connection   utils.TCPClientConnection
	writePool    sync.Pool
	writeChannel chan []byte
	doneChannel  chan bool
	Terminate    bool
	refCount     atomic.Int32
	Radars       map[string]*ClientRadar
	Metrics      hubClientMetrics
}

type hubClientMetrics struct {
	TcpClose                  *utils.Metric
	TcpOpen                   *utils.Metric
	TcpOpenError              *utils.Metric
	TcpWriteError             *utils.Metric
	TcpReadError              *utils.Metric
	TcpUnknownError           *utils.Metric
	ReadTCPBytes              *utils.Metric
	ReadTCPIterations         *utils.Metric
	WriteIncompleteError      *utils.Metric
	WriteTerminatedBytes      *utils.Metric
	WriteTerminatedIterations *utils.Metric
	WriteDequeueIterations    *utils.Metric
	WriteDequeueBytes         *utils.Metric
	TcpWriteOKIterations      *utils.Metric
	TcpWriteOKBytes           *utils.Metric
	TcpWriteErrIterations     *utils.Metric
	TcpWriteErrBytes          *utils.Metric
	ReadOKIterations          *utils.Metric
	ReadOKBytes               *utils.Metric
	ReadOverflowIterations    *utils.Metric
	ReadOverflowBytes         *utils.Metric
	ReadCorruptIterations     *utils.Metric
	ReadStarvedIterations     *utils.Metric
	ReadPopErrors             *utils.Metric
	ReadEmptyIterations       *utils.Metric
}

func (h *hubClientMetrics) init() {
	gm := &utils.GlobalMetrics

	h.TcpClose = gm.U64(hubClient, "TCP.Close.Iterations")
	h.TcpOpen = gm.U64(hubClient, "TCP.Open.Iterations")
	h.TcpOpenError = gm.U64(hubClient, "TCP.OpenError.Iterations")
	h.TcpWriteError = gm.U64(hubClient, "TCP.WritePacket.Iterations")
	h.TcpReadError = gm.U64(hubClient, "TCP.Read.Iterations")
	h.TcpUnknownError = gm.U64(hubClient, "TCP.Unknown.Iterations")
	h.ReadTCPBytes = gm.U64(hubClient, "Read.TCP.Bytes")
	h.ReadTCPIterations = gm.U64(hubClient, "Read.TCP.Iterations")
	h.WriteIncompleteError = gm.U64(hubClient, "Udp.WritePacket.Incomplete")
	h.WriteTerminatedBytes = gm.U64(hubClient, "Udp.WritePacket.Terminated.Bytes")
	h.WriteTerminatedIterations = gm.U64(hubClient, "Udp.WritePacket.Terminated.Iterations")
	h.WriteDequeueIterations = gm.U64(hubClient, "Udp.WritePacket.Dequeue.Iterations")
	h.WriteDequeueBytes = gm.U64(hubClient, "Udp.WritePacket.Dequeue.Bytes")
	h.TcpWriteOKIterations = gm.U64(hubClient, "Udp.WritePacket.OK.Iterations")
	h.TcpWriteOKBytes = gm.U64(hubClient, "Udp.WritePacket.OK.Bytes")
	h.TcpWriteErrIterations = gm.U64(hubClient, "Udp.WritePacket.Error.Iterations")
	h.TcpWriteErrBytes = gm.U64(hubClient, "Udp.WritePacket.Error.Bytes")
	h.ReadOverflowIterations = gm.U64(hubClient, "Udp.Read.Overflow.Iterations")
	h.ReadOverflowBytes = gm.U64(hubClient, "Udp.Read.Overflow.Bytes")
	h.ReadCorruptIterations = gm.U64(hubClient, "Udp.Read.Corrupt.Iterations")
	h.ReadStarvedIterations = gm.U64(hubClient, "Udp.Read.Starved.Iterations")
	h.ReadPopErrors = gm.U64(hubClient, "Udp.Read.PopError.Iterations")
	h.ReadEmptyIterations = gm.U64(hubClient, "Udp.Read.EmptyIterations")
	h.ReadOKBytes = gm.U64(hubClient, "Udp.Read.OK.Bytes")
	h.ReadOKIterations = gm.U64(hubClient, "Udp.Read.OK.Iterations")
}

func (h *Client) Start(remoteAddr utils.IP4) {
	h.writePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 2*utils.Kilobyte)
		},
	}
	h.Metrics.init()
	h.RemoteAddr = remoteAddr
	h.packetQueue.Init(4 * utils.Kilobyte)
	h.Terminate = false
	h.refCount.Store(2)
	h.writeChannel = make(chan []byte, 10)
	h.doneChannel = make(chan bool)
	h.Radars = make(map[string]*ClientRadar, 4)
	h.connection.Init(h, remoteAddr, 4, 8*utils.Kilobyte, 8*utils.Kilobyte)
	h.connection.OnError = h.onTCPError
	h.connection.OnConnect = h.onTCPOpen
	h.connection.OnDisconnect = h.onTCPClose

	go h.executeRead()
	go h.executeWrite()
}

func (h *Client) Stop() {
	if !h.Terminate {
		h.doneChannel <- true
	}
	h.Terminate = true

}

func (h *Client) StopAndJoin() {
	h.Stop()

	for h.refCount.Load() > 1 {
		time.Sleep(1 * time.Second)
	}
}

func (h *Client) executeRead() {
	var readBuffer [2 * utils.Kilobyte]byte
	var readBufferLen int
	var hubPacket tcphub.PacketWrapper

	for !h.Terminate {
		readBufferLen = h.connection.Read(readBuffer[:], time.Now(), 1000)
		now := time.Now()

		if readBufferLen > 0 {
			h.Metrics.ReadTCPBytes.Add(readBufferLen, now)
			h.Metrics.ReadTCPIterations.Inc(now)

			if h.packetQueue.GetTotalAvail() < readBufferLen {
				h.Metrics.ReadOverflowIterations.Inc(now)
				h.Metrics.ReadOverflowBytes.Add(readBufferLen+h.packetQueue.Size(), now)
			} else {
				h.Metrics.ReadOKIterations.Inc(now)
				h.Metrics.ReadOKBytes.Add(readBufferLen, now)

				h.packetQueue.PushData(readBuffer[:readBufferLen], false)

				hubPacket.Buffer = h.packetQueue.GetDataSlice()

				for hubPacket.IsParseableLength() {
					if !hubPacket.IsValidStart() {
						h.Metrics.ReadCorruptIterations.Inc(now)
						h.packetQueue.Reset()
						break
					}

					if !hubPacket.IsComplete() {
						h.Metrics.ReadStarvedIterations.Inc(now)
						break
					}

					ip4 := hubPacket.GetSourceIP4()
					ip4Str := ip4.String()

					if ip4Str != "192.168.11.2:55555" {

						// Add radar if it does not exist
						if _, ok := h.Radars[ip4Str]; !ok {
							fmt.Println("Registering", ip4Str)
							vr := &ClientRadar{
								OnUDPRead: h.onRawDataFromDesktop,
							}
							h.Radars[ip4Str] = vr
							vr.Start(ip4)
						}

						radar := h.Radars[ip4Str]
						//hubPacket.Dump("Client")
						radar.Write(hubPacket.GetPacket())

						if err := h.packetQueue.PopSize(hubPacket.GetPacketSize()); err != nil {
							h.Metrics.ReadPopErrors.Inc(now)
							h.packetQueue.Reset()
							break
						}
					}

					hubPacket.Buffer = h.packetQueue.GetDataSlice()
				}
			}
		} else {
			h.Metrics.ReadEmptyIterations.Inc(now)
		}
	}
}

func (h *Client) executeWrite() {
	for {
		select {
		case data := <-h.writeChannel:
			h.writeData(data)

		case <-h.doneChannel:
			h.Terminate = true
			h.refCount.Add(-1)
			close(h.writeChannel)
			close(h.doneChannel)
			return
		}
	}
}

func (h *Client) writeData(packetData []byte) {
	now := time.Now()

	pw := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	if h.connection.Write(pw.GetPacket(), now, 1000) {
		h.Metrics.TcpWriteOKIterations.Inc(now)
		h.Metrics.TcpWriteOKBytes.Add(pw.GetPacketSize(), now)

	} else {
		h.Metrics.TcpWriteErrIterations.Inc(now)
		h.Metrics.TcpWriteErrBytes.Add(pw.GetPacketSize(), now)
	}

	h.writePool.Put(packetData)
}

func (h *Client) Write(packetData []byte) {
	now := time.Now()

	pw := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	if !pw.IsComplete() {
		h.Metrics.WriteIncompleteError.Inc(now)
		return
	}

	if h.Terminate {
		h.Metrics.WriteTerminatedBytes.Add(pw.GetPacketSize(), now)
		h.Metrics.WriteTerminatedIterations.Inc(now)
		return
	}

	if len(h.writeChannel)+1 >= cap(h.writeChannel) {
		h.Metrics.WriteDequeueIterations.Inc(now)
		h.Metrics.WriteDequeueBytes.Add(pw.GetPacketSize(), now)
		return
	}

	packetCopy := h.writePool.Get().([]byte)
	copy(packetCopy, pw.GetPacket())
	h.writeChannel <- packetCopy
}

// onRawDataFromDesktop takes the raw data (bytes), wrap it in a packet and routes it
// to the HubServer via the Client
func (h *Client) onRawDataFromDesktop(v *ClientRadar, addr utils.IP4, bytes []byte) {
	var backingBuffer [2 * utils.Kilobyte]byte
	pw := tcphub.PacketWrapper{
		Buffer: backingBuffer[:],
	}

	fmt.Println("OnRawDataFromDesktop", v.IPAddr, addr)
	pw.Init(backingBuffer[:], 0, v.IPAddr, addr)
	pw.SetData(bytes)
	h.Write(pw.GetPacket())
}

func (h *Client) onTCPClose(_ *utils.TCPClientConnection) {
	h.Metrics.TcpClose.Inc(time.Now())
}

func (h *Client) onTCPOpen(_ *utils.TCPClientConnection) {
	h.Metrics.TcpOpen.Inc(time.Now())
}

func (h *Client) onTCPError(connection *utils.TCPClientConnection, context utils.IPErrorContext, err error) {
	switch context {
	case utils.IPErrorOnConnect:
		h.Metrics.TcpOpenError.Inc(time.Now())
	case utils.IPErrorOnReadData:
		h.Metrics.TcpReadError.Inc(time.Now())
	case utils.IPErrorOnWriteData:
		h.Metrics.TcpWriteError.Inc(time.Now())
	default:
		h.Metrics.TcpUnknownError.Inc(time.Now())
	}
}
