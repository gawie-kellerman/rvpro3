package web

import "time"

type SocketService struct {
	Terminated    bool
	PongEvery     time.Duration
	PingEvery     time.Duration
	WriteDeadline time.Duration
	MaxReadSize   int64
	clients       map[*SocketClient]bool
	register      chan *SocketClient
	unregister    chan *SocketClient
	broadcast     chan *SocketPayload
	done          chan bool
}

func NewSocketService() *SocketService {
	return &SocketService{
		clients:    make(map[*SocketClient]bool),
		register:   make(chan *SocketClient),
		unregister: make(chan *SocketClient),
		broadcast:  make(chan *SocketPayload),
	}
}

func (s *SocketService) Run() {
	s.Terminated = false
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true

		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}

		case message := <-s.broadcast:
			for client := range s.clients {
				select {
				case client.send <- message:
					// Unable to send to the client, either too slow or disconnected
				default:
					delete(s.clients, client)
					close(client.send)
				}
			}
		case <-s.done:
			s.Terminated = true
			for client := range s.clients {
				delete(s.clients, client)
				close(client.send)
			}
		}
	}
}

func (s *SocketService) NoSubscriptions(mask uint64) (res int) {
	for client := range s.clients {
		if client.Subscriptions&mask != 0 {
			res++
		}
	}
	return res
}

func (s *SocketService) Broadcast(msg *SocketPayload) {
	if s.Terminated {
		return
	}
	s.broadcast <- msg
}

func (s *SocketService) IsAnySubscribed(mask uint64) bool {
	for client := range s.clients {
		if client.Subscriptions&mask != 0 {
			return true
		}
	}
	return false
}
