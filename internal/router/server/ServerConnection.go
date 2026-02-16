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

const rsc = "Router.Server.Connection"

type HubServerConnection struct {
	connection   *net.TCPConn
	Terminate    bool
	refCount     atomic.Int32
	writePool    sync.Pool
	writeChannel chan []byte
	doneChannel  chan bool
	packetQueue  utils.QueueBuffer
	Metrics      HubServerConnectionMetrics
	OnPropagate  func(*HubServerConnection, []byte)
	OnError      func(*HubServerConnection, error)
}

type HubServerConnectionMetrics struct {
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
	utils.MetricsInitMixin
}

func (h *HubServerConnection) Start(connection *net.TCPConn) {
	h.Metrics.InitMetrics(rsc, &h.Metrics)
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
				h.Metrics.ReadEmptyIterations.IncAt(1, now)
				h.Metrics.ReadContEmptyIterations.IncAt(1, now)
			}
		} else {
			if h.packetQueue.GetTotalAvail() < bytesRead {
				h.Metrics.ReadOverflowErrors.IncAt(1, now)
				h.packetQueue.Reset()
				continue
			}
			_ = h.packetQueue.PushData(readBuffer[:bytesRead], false)

			packet := tcphub.PacketWrapper{
				Buffer: h.packetQueue.GetDataSlice(),
			}

			for packet.IsParseableLength() {
				if !packet.IsValidStart() {
					h.Metrics.ReadCorruptErrors.IncAt(1, now)
					h.packetQueue.Reset()
					break
				}

				if !packet.IsComplete() {
					h.Metrics.ReadStarvedIterations.IncAt(1, now)
					break
				}

				// Reading from a Connection means that the data should be
				// sent to a radar

				packetBytes := packet.GetPacket()
				h.onPropagate(packetBytes)

				if err := h.packetQueue.PopSize(packet.GetPacketSize()); err != nil {
					h.Metrics.ReadPopErrors.IncAt(1, now)
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
		h.Metrics.TcpWriteErrors.IncAt(1, now)
	} else {
		h.Metrics.TcpWriteIterations.IncAt(1, now)
		h.Metrics.TcpWriteBytes.IncAt(int64(packet.GetPacketSize()), now)
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
		h.Metrics.WriteIncompleteError.IncAt(1, now)
		return
	}

	if h.Terminate {
		h.Metrics.WriteTerminatedBytes.IncAt(int64(packet.GetPacketSize()), now)
		h.Metrics.WriteTerminatedIterations.IncAt(1, now)
		return
	}

	if len(h.writeChannel)+1 >= cap(h.writeChannel) {
		h.Metrics.WriteDequeueIterations.IncAt(1, now)
		h.Metrics.WriteDequeueBytes.IncAt(int64(packet.GetPacketSize()), now)
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
