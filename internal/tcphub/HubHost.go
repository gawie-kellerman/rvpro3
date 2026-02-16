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
	Metrics            HubHostMetrics
}

const hubHostMetrics = "Hub.Host"

type HubHostMetrics struct {
	TotalConnects    *utils.Metric
	TotalDisconnects *utils.Metric
	TotalRejects     *utils.Metric
	WriteToClients   *utils.Metric
	WriteBytes       *utils.Metric
	SkipToClients    *utils.Metric
	SkipBytes        *utils.Metric
	Errors           *utils.Metric
	utils.MetricsInitMixin
}

func (h *HubHost) Start(listenAddr utils.IP4) {
	h.Metrics.InitMetrics(hubHostMetrics, &h.Metrics)

	if !h.IsRunning() {
		h.Terminate = false
		h.Terminated = false

		h.ListenAddr = listenAddr

		for i := range h.Clients {
			client := &h.Clients[i]
			client.Init(h)
			client.OnConnect = func(hub *HubClient) {
				h.ClientsLen++
				h.Metrics.TotalConnects.Inc(1)
			}
			client.OnDisconnect = func(hub *HubClient) {
				h.Metrics.TotalDisconnects.Inc(1)
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
	h.Metrics.Errors.Inc(1)

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
	h.Metrics.TotalRejects.Inc(1)

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
		h.Metrics.WriteToClients.IncAt(1, now)
		h.Metrics.WriteBytes.IncAt(1, now)

		for i := range h.Clients {
			client := &h.Clients[i]
			client.Write(packet)
		}
	} else {
		h.Metrics.SkipToClients.IncAt(1, now)
		h.Metrics.SkipBytes.IncAt(int64(packet.Size), now)
	}
}
