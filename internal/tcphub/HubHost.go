package tcphub

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

var GlobalHubHost HubHost

type HubHost struct {
	Terminate          bool
	Terminated         bool
	ListenAddr         utils.IP4
	Clients            [2]HubClient
	ClientsLen         int
	OnError            func(*HubHost, error)        `json:"-"`
	OnRejectConnection func(*HubHost, *net.TCPConn) `json:"-"`
	OnAcceptConnection func(*HubHost, *net.TCPConn) `json:"-"`
	doneChan           chan bool                    `json:"-"`
	dispatcher         HubDispatcher                `json:"-"`
	metrics            metrics
}

const hubHostMetrics = "Hub.Host"

type metrics struct {
	totalConnects    *utils.Metric
	totalDisconnects *utils.Metric
	totalRejects     *utils.Metric
	writeToClients   *utils.Metric
	writeBytes       *utils.Metric
	skipToClients    *utils.Metric
	skipBytes        *utils.Metric
	errors           *utils.Metric
}

func (m *metrics) init() {
	gm := &utils.GlobalMetrics
	m.totalConnects = gm.U64(hubHostMetrics, "Total Connects")
	m.totalDisconnects = gm.U64(hubHostMetrics, "Total Disconnects")
	m.totalRejects = gm.U64(hubHostMetrics, "Total Connection Rejects")
	m.writeToClients = gm.U64(hubHostMetrics, "WritePacket to Clients")
	m.writeBytes = gm.U64(hubHostMetrics, "WritePacket bytes")
	m.skipToClients = gm.U64(hubHostMetrics, "Skip to Clients")
	m.errors = gm.U64(hubHostMetrics, "Errors")
}

func (h *HubHost) Start(listenAddr utils.IP4) {
	h.metrics.init()

	if !h.IsRunning() {
		h.Terminate = false
		h.Terminated = false

		h.ListenAddr = listenAddr

		for i := range h.Clients {
			client := &h.Clients[i]
			client.Init(h)
			client.OnConnect = func(hub *HubClient) {
				h.ClientsLen++
				h.metrics.totalConnects.Inc(time.Now())
			}
			client.OnDisconnect = func(hub *HubClient) {
				h.metrics.totalDisconnects.Inc(time.Now())
				h.ClientsLen--
			}
		}
		go h.executeListen()

		h.dispatcher.Start(listenAddr)
	}
}

func (h *HubHost) IsRunning() bool {
	return h.Terminated
}

func (h *HubHost) Stop() {
	for i := range h.Clients {
		client := &h.Clients[i]
		client.Stop()
	}

	h.Terminate = true

	for !h.Terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

// executeListen listens for client connections
func (h *HubHost) executeListen() {
	var listener *net.TCPListener
	var err error
	addr := h.ListenAddr.ToTCPAddr()

	for !h.Terminate {
		if listener == nil {
			if listener, err = net.ListenTCP("tcp", &addr); err != nil {
				listener = nil
				h.onError(err)
				time.Sleep(time.Second)
			}
		}

		if listener != nil {
			var conn *net.TCPConn
			_ = listener.SetDeadline(time.Now().Add(time.Second * 3))
			if conn, err = listener.AcceptTCP(); err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					h.onError(err)
					_ = listener.Close()
					listener = nil
					time.Sleep(time.Second)
				}
			} else {
				index := h.GetFreeSlot()

				if index == -1 {
					h.onRejectConnection(conn)
					_ = conn.Close()
				} else {
					h.onAcceptConnection(conn)
					hubClient := &h.Clients[index]
					hubClient.Start(*conn)
				}
			}
		}
	}

	if listener != nil {
		_ = listener.Close()
	}
}

func (h *HubHost) onError(err error) {
	h.metrics.errors.Inc(time.Now())

	if h.OnError != nil {
		h.OnError(h, err)
	} else {
		log.Err(err).Msg("HubHost.onError")
	}
}

func (h *HubHost) GetFreeSlot() int {
	for i := range h.Clients {
		client := &h.Clients[i]
		if client.IsTerminated() {
			return i
		}
	}
	return -1
}

func (h *HubHost) onRejectConnection(client *net.TCPConn) {
	h.metrics.totalRejects.Inc(time.Now())

	if h.OnRejectConnection != nil {
		h.OnRejectConnection(h, client)
	}
}

func (h *HubHost) onAcceptConnection(client *net.TCPConn) {
	if h.OnAcceptConnection != nil {
		h.OnAcceptConnection(h, client)
	}
}

func (h *HubHost) WriteToClients(packet Packet) {
	now := time.Now()
	if h.ClientsLen > 0 {
		h.metrics.writeToClients.Inc(now)
		h.metrics.writeBytes.Inc(now)

		for i := range h.Clients {
			client := &h.Clients[i]
			client.Write(packet)
		}
	} else {
		h.metrics.skipToClients.Inc(now)
		h.metrics.skipBytes.Add(int(packet.Size), now)
	}
}
