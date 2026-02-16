package client

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

type ClientRadar struct {
	IPAddr         utils.IP4 `json:"IPAddr"`
	Terminate      bool      `json:"Terminate"`
	writePool      sync.Pool
	writeChannel   chan []byte
	doneChannel    chan bool
	refCount       atomic.Int32
	connection     utils.UDPServerConnection
	now            time.Time
	Metrics        ClientRadarMetrics
	OnUDPRead      func(v *ClientRadar, addr utils.IP4, bytes []byte)
	OnWriteSuccess func(v *ClientRadar, targetIP utils.IP4, dataOnly []byte)
	OnWriteFail    func(v *ClientRadar, targetIP utils.IP4, dataOnly []byte, err error)
}

type ClientRadarMetrics struct {
	UdpOpen                   *utils.Metric
	UdpClose                  *utils.Metric
	UdpOpenError              *utils.Metric
	UdpWriteError             *utils.Metric
	UdpReadError              *utils.Metric
	UdpUnknownError           *utils.Metric
	UdpReadBytes              *utils.Metric
	UdpReadIterations         *utils.Metric
	UdpWATBytes               *utils.Metric //  WritePacket After Terminate
	UdpWATIterations          *utils.Metric
	WriteIncompleteError      *utils.Metric
	WriteTerminatedBytes      *utils.Metric
	WriteTerminatedIterations *utils.Metric
	WriteDequeueIterations    *utils.Metric
	WriteDequeueBytes         *utils.Metric
	UdpWriteOKIterations      *utils.Metric
	UdpWriteOKBytes           *utils.Metric
	UdpWriteErrIterations     *utils.Metric
	UdpWriteErrBytes          *utils.Metric
	utils.MetricsInitMixin
}

func (m *ClientRadarMetrics) inits(addr utils.IP4) {
}

func (v *ClientRadar) Start(ipAddr utils.IP4) {
	v.writePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 2*utils.Kilobyte)
		},
	}
	v.IPAddr = ipAddr
	v.Terminate = false
	v.refCount.Store(2)
	v.writeChannel = make(chan []byte, 10) //TODO: Externalize the write message queue size
	v.doneChannel = make(chan bool)

	sectionName := "Router.Client.Radar.[" + ipAddr.String() + "]"
	v.Metrics.InitMetrics(sectionName, &v.Metrics)

	v.connection.Init(v, ipAddr, 2*utils.Kilobyte, 2*utils.Kilobyte, 3) // TODO: Externalize
	v.connection.OnError = v.onUDPError
	v.connection.OnOpen = v.onUDPOpen
	v.connection.OnClose = v.onUDPClose

	// Attempt to open before using it
	v.connection.Listen()

	go v.executeRead()
	go v.executeWrite()
}

func (v *ClientRadar) Stop() {
	if !v.Terminate {
		v.doneChannel <- true
	}
	v.Terminate = true

	for v.refCount.Load() > 1 {
		time.Sleep(100 * time.Millisecond)
	}
}

// executeRead reads instructions sent from Traffic UI to the radar on the local desktop
func (v *ClientRadar) executeRead() {
	var buffer [2 * utils.Kilobyte]byte
	var bufferLen int
	var fromAddr utils.IP4

	for !v.Terminate {
		v.now = time.Now()
		if cnx := v.connection.Listen(); cnx != nil {
			bufferLen, fromAddr = v.connection.Read(buffer[:], v.now, 1000) // TODO: Externalize

			if bufferLen > 0 {
				fmt.Println("Reading from ", fromAddr.String(), "bytes", bufferLen)
				v.Metrics.UdpReadBytes.IncAt(int64(bufferLen), v.now)
				v.Metrics.UdpReadIterations.IncAt(1, v.now)
				if v.OnUDPRead != nil {
					v.OnUDPRead(v, fromAddr, buffer[:bufferLen])
				}
			}
		} else {
			time.Sleep(time.Millisecond * 1000)
		}
	}

	v.refCount.Add(-1)
}

// executeWrite writes data from the remote radar to the local desktop
func (v *ClientRadar) executeWrite() {
	for {
		select {
		case data := <-v.writeChannel:
			v.writeData(data)

		case <-v.doneChannel:
			v.Terminate = true
			v.refCount.Add(-1)
			v.connection.Close()

			close(v.writeChannel)
			close(v.doneChannel)
			return
		}
	}
}

// writeData interprets and write HubPacket data, as the Virtual Radar to Traffic UI
// The Write method already validated the packetData
func (v *ClientRadar) writeData(packetData []byte) {
	var targetAddr utils.IP4
	if !v.Terminate {
		now := time.Now()

		pw := tcphub.PacketWrapper{
			Buffer: packetData,
		}

		//pw.Dump("Radar_")

		targetType := pw.GetPacketType()

		if targetType == tcphub.PtRadarMulticast {
			// Sending the multicast
			targetAddr = utils.IP4Builder.FromString("239.144.0.0:60000")
			fmt.Println("Writing multicast from ", v.IPAddr, "to", targetAddr.String(), "bytes")
		} else {
			targetAddr = pw.GetTargetIP4()
		}
		if err := v.connection.WriteData(targetAddr.ToUDPAddr(), pw.GetData()); err == nil {
			v.Metrics.UdpWriteOKIterations.IncAt(1, now)
			v.Metrics.UdpWriteOKBytes.IncAt(int64(pw.GetDataSize()), now)

			if v.OnWriteSuccess != nil {
				v.OnWriteSuccess(v, pw.GetTargetIP4(), pw.GetData())
			}
		} else {
			v.Metrics.UdpWriteErrIterations.IncAt(1, now)
			v.Metrics.UdpWriteErrBytes.IncAt(int64(pw.GetDataSize()), now)

			if v.OnWriteFail != nil {
				v.OnWriteFail(v, pw.GetTargetIP4(), pw.GetData(), err)
			}
		}
	}

	v.writePool.Put(packetData)
}

// Write writes HubPacket data to desktop.  The data buffer passed is likely
// bigger than the packet to be written, meaning the data need to be trimmed
// to the packet size
func (v *ClientRadar) Write(packetData []byte) {
	now := time.Now()

	pw := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	if !pw.IsComplete() {
		v.Metrics.WriteIncompleteError.IncAt(1, now)
		return
	}

	if v.Terminate {
		v.Metrics.WriteTerminatedBytes.IncAt(int64(pw.GetPacketSize()), now)
		v.Metrics.WriteTerminatedIterations.IncAt(1, now)
		return
	}

	if len(v.writeChannel)+1 >= cap(v.writeChannel) {
		v.Metrics.WriteDequeueIterations.IncAt(1, now)
		v.Metrics.WriteDequeueBytes.IncAt(int64(pw.GetPacketSize()), now)
		return
	}

	packetCopy := v.writePool.Get().([]byte)
	copy(packetCopy, pw.GetPacket())

	v.writeChannel <- packetCopy
}

func (v *ClientRadar) IsTerminated() bool {
	return v.refCount.Load() == 0
}

func (v *ClientRadar) onUDPError(
	_ *utils.UDPServerConnection,
	context utils.IPErrorContext,
	_ error,
) {
	switch context {
	case utils.IPErrorOnConnect:
		v.Metrics.UdpOpenError.Inc(1)
	case utils.IPErrorOnReadData:
		v.Metrics.UdpReadError.Inc(1)
	case utils.IPErrorOnWriteData:
		v.Metrics.UdpWriteError.Inc(1)
	default:
		v.Metrics.UdpUnknownError.Inc(1)
	}

	// TODO: Add callback to desktop listening in...
}

func (v *ClientRadar) onUDPOpen(connection *utils.UDPServerConnection) {
	v.Metrics.UdpOpen.Inc(1)
}

func (v *ClientRadar) onUDPClose(connection *utils.UDPServerConnection) {
	v.Metrics.UdpClose.Inc(1)
}
