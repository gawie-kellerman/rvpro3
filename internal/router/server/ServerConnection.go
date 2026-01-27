package server

import (
	"bufio"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

const hsc = "Hub.Server.Connection"

type HubServerConnection struct {
	connection   *net.TCPConn
	Terminate    bool
	refCount     atomic.Int32
	writePool    sync.Pool
	writeChannel chan []byte
	doneChannel  chan bool
	packetQueue  utils.QueueBuffer
	Metrics      hubServerConnectionMetrics
	OnPropagate  func(*HubServerConnection, []byte)
	OnError      func(*HubServerConnection, error)
}

type hubServerConnectionMetrics struct {
	ReadOverflowErrors        *utils.Metric
	ReadEmptyIterations       *utils.Metric
	ReadContEmptyIterations   *utils.Metric
	ReadCorruptErrors         *utils.Metric
	ReadStarvedIterations     *utils.Metric
	ReadPopErrors             *utils.Metric
	TcpWriteErrors            *utils.Metric
	TcpWriteIterations        *utils.Metric
	TcpWriteBytes             *utils.Metric
	WriteIncompleteError      *utils.Metric
	WriteTerminatedBytes      *utils.Metric
	WriteTerminatedIterations *utils.Metric
	WriteDequeueIterations    *utils.Metric
	WriteDequeueBytes         *utils.Metric
}

func (h *hubServerConnectionMetrics) init() {
	gm := &utils.GlobalMetrics
	h.ReadContEmptyIterations = gm.U64(hsc, "Read.ContEmpty.Iterations")
	h.ReadOverflowErrors = gm.U64(hsc, "Read.Overflow.Errors")
	h.ReadEmptyIterations = gm.U64(hsc, "Read.Empty.Iterations")
	h.ReadCorruptErrors = gm.U64(hsc, "Read.Corrupt.Errors")
	h.ReadStarvedIterations = gm.U64(hsc, "Read.Starved.Iterations")
	h.ReadPopErrors = gm.U64(hsc, "Read.Pop.Errors")
	h.TcpWriteErrors = gm.U64(hsc, "TCP.WritePacket.Errors")
	h.TcpWriteIterations = gm.U64(hsc, "TCP.WritePacket.Iterations")
	h.TcpWriteBytes = gm.U64(hsc, "TCP.WritePacket.Bytes")
	h.WriteIncompleteError = gm.U64(hsc, "WritePacket.Incomplete.Errors")
	h.WriteTerminatedBytes = gm.U64(hsc, "WritePacket.Terminated.Bytes")
	h.WriteTerminatedIterations = gm.U64(hsc, "WritePacket.Terminated.Iterations")
	h.WriteDequeueIterations = gm.U64(hsc, "WritePacket.Dequeue.Iterations")
	h.WriteDequeueBytes = gm.U64(hsc, "WritePacket.Dequeue.Bytes")
}

func (h *HubServerConnection) Start(connection *net.TCPConn) {
	h.Metrics.init()
	h.connection = connection
	h.Terminate = false
	h.refCount.Store(2)
	h.writeChannel = make(chan []byte, 10)
	h.doneChannel = make(chan bool)
	h.packetQueue.Init(8 * utils.Kilobyte)
	h.writePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 2*utils.Kilobyte)
		},
	}

	go h.executeRead()
	go h.executeWrite()
}

func (h *HubServerConnection) Stop() {
	if !h.Terminate {
		h.Terminate = true
		h.doneChannel <- true
	}

	if h.connection != nil {
		h.connection.Close()
	}
	h.connection = nil
}

func (h *HubServerConnection) StopAndJoin() {
	h.Stop()

	for h.refCount.Load() > 1 {
		time.Sleep(100 * time.Millisecond)
	}
}

// executeRead reads data sent from HubClient to Server which is in all likelihood
// Traffic UI instructions (to the radar)
func (h *HubServerConnection) executeRead() {
	var readBuffer [4 * utils.Kilobyte]byte
	reader := bufio.NewReader(h.connection)

	for !h.Terminate {
		now := time.Now()
		_ = h.connection.SetReadDeadline(now.Add(1 * time.Second))
		bytesRead, err := reader.Read(readBuffer[:])

		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				h.onError(err)
				h.Stop()
				break
			} else {
				h.Metrics.ReadEmptyIterations.Inc(now)
				h.Metrics.ReadContEmptyIterations.Inc(now)
			}
		} else {
			if h.packetQueue.GetTotalAvail() < bytesRead {
				h.Metrics.ReadOverflowErrors.Inc(now)
				h.packetQueue.Reset()
				continue
			}
			_ = h.packetQueue.PushData(readBuffer[:bytesRead], false)

			packet := tcphub.PacketWrapper{
				Buffer: h.packetQueue.GetDataSlice(),
			}

			for packet.IsParseableLength() {
				if !packet.IsValidStart() {
					h.Metrics.ReadCorruptErrors.Inc(now)
					h.packetQueue.Reset()
					break
				}

				if !packet.IsComplete() {
					h.Metrics.ReadStarvedIterations.Inc(now)
					break
				}

				// Reading from a Connection means that the data should be
				// sent to a radar

				packetBytes := packet.GetPacket()
				h.onPropagate(packetBytes)

				if err := h.packetQueue.PopSize(packet.GetPacketSize()); err != nil {
					h.Metrics.ReadPopErrors.Inc(now)
					h.packetQueue.Reset()
					break
				}

				packet.Buffer = h.packetQueue.GetDataSlice()
			}
		}
	}

	h.refCount.Add(-1)
}

func (h *HubServerConnection) executeWrite() {
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

// writeData writes Server radar data to the connected HubClient
func (h *HubServerConnection) writeData(packetData []byte) {
	connection := h.connection
	if connection == nil {
		return
	}

	now := time.Now()

	packet := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	_ = connection.SetWriteDeadline(now.Add(3 * time.Second))
	_, err := h.connection.Write(packet.GetPacket())

	if err != nil {
		h.onError(err)
		h.Stop()
		h.Metrics.TcpWriteErrors.Inc(now)
	} else {
		h.Metrics.TcpWriteIterations.Inc(now)
		h.Metrics.TcpWriteBytes.Add(packet.GetPacketSize(), now)
		//packet.Dump("Server")
	}

	h.writePool.Put(packetData)
}

func (h *HubServerConnection) Write(packetData []byte) {
	now := time.Now()

	packet := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	if !packet.IsComplete() {
		h.Metrics.WriteIncompleteError.Inc(now)
		return
	}

	if h.Terminate {
		h.Metrics.WriteTerminatedBytes.Add(packet.GetPacketSize(), now)
		h.Metrics.WriteTerminatedIterations.Inc(now)
		return
	}

	if len(h.writeChannel)+1 >= cap(h.writeChannel) {
		h.Metrics.WriteDequeueIterations.Inc(now)
		h.Metrics.WriteDequeueBytes.Add(packet.GetPacketSize(), now)
		return
	}

	packetCopy := h.writePool.Get().([]byte)
	copy(packetCopy, packet.GetPacket())
	h.writeChannel <- packetData
}

func (h *HubServerConnection) onPropagate(bytes []byte) {
	if h.OnPropagate != nil {
		h.OnPropagate(h, bytes)
	}
}

func (h *HubServerConnection) onError(err error) {
	if h.OnError != nil {
		h.OnError(h, err)
	} else {
		log.Err(err).Msgf("HubServerConnection.onError remote: %s, local: %s", h.connection.RemoteAddr(), h.connection.LocalAddr())
	}
}

func (h *HubServerConnection) IsTerminated() bool {
	return h.refCount.Load() == 0
}
