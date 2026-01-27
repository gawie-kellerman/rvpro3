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

const hubServer = "Hub.Server"

type Server struct {
	BindAddr     utils.IP4
	Metrics      hubServerMetrics
	Terminate    bool
	Writer       interfaces.IUDPPacketWriter
	writeChannel chan []byte
	doneChannel  chan bool
	refCount     atomic.Int32
	OnError      func(*Server, error)
	Connections  map[string]*HubServerConnection
}

type hubServerMetrics struct {
	ListenerStartErrors  *utils.Metric
	ListenerAcceptErrors *utils.Metric
	PropagateErrors      *utils.Metric
	PropagateErrorBytes  *utils.Metric
	PropagateOKs         *utils.Metric
	PropagateOKBytes     *utils.Metric
	PropagateNoops       *utils.Metric
	PropagateNoopBytes   *utils.Metric
}

func (h *hubServerMetrics) init() {
	gm := &utils.GlobalMetrics

	h.ListenerStartErrors = gm.U64(hubServer, "Listener.Start.Errors")
	h.ListenerAcceptErrors = gm.U64(hubServer, "Listener.Accept.Errors")
	h.PropagateErrors = gm.U64(hubServer, "Listener.Propagate.Errors")
	h.PropagateErrorBytes = gm.U64(hubServer, "Listener.Propagate.Errors.Bytes")
	h.PropagateOKs = gm.U64(hubServer, "Listener.Propagate.OK")
	h.PropagateOKBytes = gm.U64(hubServer, "Listener.Propagate.OK.Bytes")
	h.PropagateNoops = gm.U64(hubServer, "Listener.Propagate.Noop")
	h.PropagateNoopBytes = gm.U64(hubServer, "Listener.Propagate.Noop.Bytes")
}

func (h *Server) Start(bindAddr utils.IP4, writer interfaces.IUDPPacketWriter) {
	h.Metrics.init()
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
				h.Metrics.ListenerStartErrors.Inc(now)
				h.onError(err)
				time.Sleep(time.Second)
			}
		}

		if listener != nil {
			var conn *net.TCPConn
			_ = listener.SetDeadline(now.Add(time.Second * 1))

			if conn, err = listener.AcceptTCP(); err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					h.Metrics.ListenerAcceptErrors.Inc(now)

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
			h.Metrics.PropagateErrors.Inc(now)
			h.Metrics.PropagateErrorBytes.Add(packet.GetPacketSize(), now)
		} else {
			h.Metrics.PropagateOKs.Inc(now)
			h.Metrics.PropagateOKBytes.Add(packet.GetPacketSize(), now)
		}
	} else {
		h.Metrics.PropagateNoops.Inc(now)
		h.Metrics.PropagateNoopBytes.Add(packet.GetPacketSize(), now)
	}
}
