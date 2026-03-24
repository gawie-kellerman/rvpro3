package server

import (
	"bufio"
	"fmt"
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

type RouterServerConnection struct {
	connection     *net.TCPConn
	ReadTerminate  bool
	WriteTerminate bool
	refCount       atomic.Int32
	writePool      sync.Pool
	writeChannel   chan []byte
	doneChannel    chan bool
	packetQueue    utils.QueueBuffer
	Metrics        HubServerConnectionMetrics
	OnPropagate    func(*RouterServerConnection, []byte)
	OnError        func(*RouterServerConnection, error)
	OnClose        func(*RouterServerConnection, error)
	DoneTerminate  bool
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

func (h *RouterServerConnection) Start(connection *net.TCPConn) {
	h.Metrics.InitMetrics(rsc, &h.Metrics)
	h.connection = connection
	h.ReadTerminate = false
	h.refCount.Store(2)
	h.writeChannel = make(chan []byte, 10)
	h.doneChannel = make(chan bool)
	h.packetQueue.Init(8 * utils.Kilobyte)
	h.writePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 2*utils.Kilobyte)
		},
	}

	fmt.Println("onstart")
	go h.executeRead()
	go h.executeWrite()
	go h.executeError()
}

func (h *RouterServerConnection) Stop() {
	if !h.ReadTerminate && h.doneChannel != nil {
		h.ReadTerminate = true
		h.doneChannel <- true
	}

	if h.connection != nil {
		h.connection.Close()
	}
	h.connection = nil
}

func (h *RouterServerConnection) Join() {
	for h.refCount.Load() > 1 {
		time.Sleep(100 * time.Millisecond)
	}

	if h.OnClose != nil {
		fmt.Println("close OnClose")
		h.OnClose(h, nil)
	}
}

// executeRead reads data sent from HubClient to Server which is in all likelihood
// Traffic UI instructions (to the radar)
func (h *RouterServerConnection) executeRead() {
	var readBuffer [4 * utils.Kilobyte]byte
	reader := bufio.NewReader(h.connection)

	for !h.ReadTerminate {
		now := time.Now()
		_ = h.connection.SetReadDeadline(now.Add(1 * time.Second))
		bytesRead, err := reader.Read(readBuffer[:])

		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				h.onError(err)
				if !h.DoneTerminate {
					h.DoneTerminate = true
					h.doneChannel <- true
				}
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

func (h *RouterServerConnection) executeWrite() {
	for msg := range h.writeChannel {
		if h.writeData(msg) != nil {
			if !h.DoneTerminate {
				h.DoneTerminate = true
				h.doneChannel <- true
			}
		}

		//case <-h.doneChannel:
		//	h.WriteTerminate = true
		//	h.ReadTerminate = true
		//
		//	h.refCount.Add(-1)
		//
		//	close(h.writeChannel)
		//	close(h.doneChannel)
		//
		//	h.Join()
		//	return
		//}
	}
}

// writeData writes Server radar data to the connected HubClient
func (h *RouterServerConnection) writeData(packetData []byte) error {
	if h.WriteTerminate {
		return nil
	}
	connection := h.connection

	if connection == nil {
		return nil
	}

	now := time.Now()

	packet := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	_ = connection.SetWriteDeadline(now.Add(3 * time.Second))
	_, err := h.connection.Write(packet.GetPacket())

	if err != nil {
		h.onError(err)
		h.Metrics.TcpWriteErrors.IncAt(1, now)
	} else {
		h.Metrics.TcpWriteIterations.IncAt(1, now)
		h.Metrics.TcpWriteBytes.IncAt(int64(packet.GetPacketSize()), now)
	}

	h.writePool.Put(packetData)
	return err
}

func (h *RouterServerConnection) Write(packetData []byte) {
	if h.WriteTerminate {
		return
	}

	now := time.Now()

	packet := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	if !packet.IsComplete() {
		h.Metrics.WriteIncompleteError.IncAt(1, now)
		return
	}

	if h.ReadTerminate || h.writeChannel == nil {
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

func (h *RouterServerConnection) onPropagate(bytes []byte) {
	if h.OnPropagate != nil {
		h.OnPropagate(h, bytes)
	}
}

func (h *RouterServerConnection) onError(err error) {
	if h.OnError != nil {
		h.OnError(h, err)
	} else {
		log.Err(err).Msgf("RouterServerConnection.onError remote: %s, local: %s", h.connection.RemoteAddr(), h.connection.LocalAddr())
	}
}

func (h *RouterServerConnection) IsTerminated() bool {
	return h.connection == nil || h.refCount.Load() == 0
}

func (h *RouterServerConnection) executeError() {
	for range h.doneChannel {

		h.WriteTerminate = true
		h.ReadTerminate = true

		h.refCount.Add(-1)

		close(h.writeChannel)
		close(h.doneChannel)

		h.Join()
		return
	}
}
