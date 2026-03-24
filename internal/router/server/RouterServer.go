package server

import (
	"fmt"
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

type RouterServer struct {
	MultiAddr      utils.IP4
	BindAddr       utils.IP4
	Terminate      bool
	Metrics        ServerMetrics
	MulticastError error
	writeChannel   chan []byte
	doneChannel    chan bool
	refCount       atomic.Int32
	Writer         interfaces.IUDPWriter              `json:"-"`
	OnError        func(*RouterServer, error)         `json:"-"`
	Connections    map[string]*RouterServerConnection `json:"-"`
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

func (h *RouterServer) Start(
	bindAddr utils.IP4,
	multiAddr utils.IP4,
	writer interfaces.IUDPWriter) {
	h.Metrics.InitMetrics(routerServer, &h.Metrics)
	h.Connections = make(map[string]*RouterServerConnection)
	h.BindAddr = bindAddr
	h.MultiAddr = multiAddr
	h.Terminate = false
	h.Writer = writer
	h.refCount.Store(2)

	go h.executeAccept()
	go h.executeMulticast()
}

func (h *RouterServer) executeMulticast() {
	multiAddr, err := net.ResolveUDPAddr("udp", h.MultiAddr.String())
	if err != nil {
		log.Error().Err(err).Msg("Error resolving multicast address")
		return
	}

	var conn *net.UDPConn
	var buffer [60]byte

	for !h.Terminate {
		if conn == nil {
			conn, err = net.ListenMulticastUDP("udp", nil, multiAddr)
			if err != nil {
				if !errors.Is(err, h.MulticastError) {
					log.Error().Err(err).Msg("Error listening multicast UDP")
				}
				time.Sleep(3 * time.Second)
				h.MulticastError = err
			} else {
				_ = conn.SetReadBuffer(cap(buffer))
			}

		}

		if conn != nil {
			readBytes, readAddr, readErr := conn.ReadFromUDP(buffer[:])

			if readErr != nil {
				log.Error().Err(readErr).Msg("Error reading multicast packet")
				_ = conn.Close()
				conn = nil
			} else {
				remoteIP := utils.IP4Builder.FromAddr(readAddr)

				if remoteIP.Port == 60000 {
					var packetBuffer [256]byte
					var packet tcphub.PacketWrapper
					packet.Init(packetBuffer[:], 0, h.MultiAddr, remoteIP)
					packet.SetPacketType(tcphub.PtRadarMulticast)
					packet.SetData(buffer[:readBytes])
					h.Write(packet.GetPacket())
				}
			}
		}
	}

	if conn != nil {
		_ = conn.Close()
	}
}

func (h *RouterServer) executeAccept() {
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

func (h *RouterServer) StartClient(conn *net.TCPConn) {
	h.clearConnections()

	ip4 := utils.IP4Builder.FromAddr(conn.RemoteAddr())
	obj := &RouterServerConnection{}
	obj.OnPropagate = h.onPropagateConnectionData
	obj.OnClose = h.onCloseConnection
	h.Connections[ip4.String()] = obj
	obj.Start(conn)
}

// Write sends the packet data to each client.  Each client determines
// its backlog, and copy the packet if it thinks possible to send it
func (h *RouterServer) Write(packetData []byte) {
	for _, conn := range h.Connections {
		conn.Write(packetData)
	}
}

func (h *RouterServer) onError(err error) {
	if h.OnError != nil {
		h.OnError(h, err)
	} else {
		log.Err(err).Msg("Hub Server Error")
	}
}

func (h *RouterServer) clearConnections() {

	for key, conn := range h.Connections {
		if conn.IsTerminated() {
			delete(h.Connections, key)
		}
	}
	fmt.Println("Clear Connections", len(h.Connections))
}

// onPropagateConnectionData takes the data as sent from the connection and send it
// to the Server associated packet writer (generally udp sent to the radar)
func (h *RouterServer) onPropagateConnectionData(connection *RouterServerConnection, packetData []byte) {
	now := time.Now()

	packet := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	if h.Writer != nil {
		if err := h.Writer.WriteData(packet.GetTargetIP4(), packet.GetData()); err != nil {
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

func (h *RouterServer) HasConnections() bool {
	return len(h.Connections) > 0
}

func (h *RouterServer) onCloseConnection(*RouterServerConnection, error) {
	h.clearConnections()
}
