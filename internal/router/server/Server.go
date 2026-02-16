package server

import (
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

const routerServer = "Router.Server"

type Server struct {
	BindAddr     utils.IP4
	Terminate    bool
	Writer       interfaces.IUDPPacketWriter
	Metrics      ServerMetrics
	writeChannel chan []byte
	doneChannel  chan bool
	refCount     atomic.Int32
	OnError      func(*Server, error)
	Connections  map[string]*HubServerConnection
}

type ServerMetrics struct {
	ListenerStartErrors  *utils.Metric
	ListenerAcceptErrors *utils.Metric
	PropagateErrors      *utils.Metric
	PropagateErrorBytes  *utils.Metric
	PropagateOKs         *utils.Metric
	PropagateOKBytes     *utils.Metric
	PropagateNoops       *utils.Metric
	PropagateNoopBytes   *utils.Metric
	utils.MetricsInitMixin
}

func (h *Server) Start(bindAddr utils.IP4, writer interfaces.IUDPPacketWriter) {
	h.Metrics.InitMetrics(routerServer, &h.Metrics)
	h.Connections = make(map[string]*HubServerConnection)
	h.BindAddr = bindAddr
	h.Terminate = false
	h.Writer = writer
	h.refCount.Store(2)

	go h.executeAccept()
}

func (h *Server) executeAccept() {
	var listener *net.TCPListener
	var err error

	addr := h.BindAddr.ToTCPAddr()

	for !h.Terminate {
		now := time.Now()

		if listener == nil {
			if listener, err = net.ListenTCP("tcp", &addr); err != nil {
				listener = nil
				h.Metrics.ListenerStartErrors.IncAt(1, now)
				h.onError(err)
				time.Sleep(time.Second)
			}
		}

		if listener != nil {
			var conn *net.TCPConn
			_ = listener.SetDeadline(now.Add(time.Second * 1))

			if conn, err = listener.AcceptTCP(); err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					h.Metrics.ListenerAcceptErrors.IncAt(1, now)

					_ = listener.Close()
					listener = nil

					h.onError(err)
				}
			} else {
				// No error occurred
				h.StartClient(conn)
			}
		}
	}

	if listener != nil {
		_ = listener.Close()
	}

	h.refCount.Add(-1)
}

func (h *Server) StartClient(conn *net.TCPConn) {
	h.clearConnections()

	ip4 := utils.IP4Builder.FromAddr(conn.RemoteAddr())
	obj := &HubServerConnection{}
	obj.OnPropagate = h.onPropagateConnectionData
	h.Connections[ip4.String()] = obj
	obj.Start(conn)
}

// Write sends the packet data to each client.  Each client determines
// its backlog, and copy the packet if it thinks possible to send it
func (h *Server) Write(packetData []byte) {
	for _, conn := range h.Connections {
		conn.Write(packetData)
	}
}

func (h *Server) onError(err error) {
	if h.OnError != nil {
		h.OnError(h, err)
	} else {
		log.Err(err).Msg("Hub Server Error")
	}
}

func (h *Server) clearConnections() {
	for key, conn := range h.Connections {
		if conn.IsTerminated() {
			delete(h.Connections, key)
		}
	}
}

// onPropagateConnectionData takes the data as sent from the connection and send it
// to the Server associated packet writer (generally udp sent to the radar)
func (h *Server) onPropagateConnectionData(connection *HubServerConnection, packetData []byte) {
	now := time.Now()

	packet := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	if h.Writer != nil {
		if err := h.Writer.WritePacket(packetData); err != nil {
			h.Metrics.PropagateErrors.IncAt(1, now)
			h.Metrics.PropagateErrorBytes.IncAt(int64(packet.GetPacketSize()), now)
		} else {
			h.Metrics.PropagateOKs.IncAt(1, now)
			h.Metrics.PropagateOKBytes.IncAt(int64(packet.GetPacketSize()), now)
		}
	} else {
		h.Metrics.PropagateNoops.IncAt(1, now)
		h.Metrics.PropagateNoopBytes.IncAt(int64(packet.GetPacketSize()), now)
	}
}
