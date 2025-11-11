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
	terminate  bool
	terminated bool
	listenAddr utils.IP4
	doneChan   chan bool
	clients    [2]HubClient
	dispatcher HubDispatcher

	OnError            func(*HubHost, error)
	OnRejectConnection func(*HubHost, *net.TCPConn)
	OnAcceptConnection func(*HubHost, *net.TCPConn)
}

func (h *HubHost) Start(listenAddr utils.IP4) {
	if !h.IsRunning() {
		h.terminate = false
		h.terminated = false

		h.listenAddr = listenAddr

		for i := range h.clients {
			client := &h.clients[i]
			client.Init(h)
		}
		go h.executeListen()

		h.dispatcher.Start(listenAddr)
	}
}

func (h *HubHost) IsRunning() bool {
	return h.terminated
}

func (h *HubHost) Stop() {
	for i := range h.clients {
		client := &h.clients[i]
		client.Stop()
	}

	h.terminate = true

	for !h.terminated {
		time.Sleep(1 * time.Second)
	}
}

func (h *HubHost) executeListen() {
	var listener *net.TCPListener
	var err error
	addr := h.listenAddr.ToTCPAddr()

	for !h.terminate {
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
					hubClient := &h.clients[index]
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
	if h.OnError != nil {
		h.OnError(h, err)
	} else {
		log.Err(err).Msg("HubHost.onError")
	}
}

func (h *HubHost) GetFreeSlot() int {
	for i := range h.clients {
		client := &h.clients[i]
		if client.IsTerminated() {
			return i
		}
	}
	return -1
}

func (h *HubHost) onRejectConnection(client *net.TCPConn) {
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
	for i := range h.clients {
		client := &h.clients[i]
		client.Write(packet)
	}

}
