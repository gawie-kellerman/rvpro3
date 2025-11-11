package tcphub

import (
	"net"

	"github.com/rs/zerolog/log"

	"rvpro3/radarvision.com/utils"
)

var GlobalTCPHubServer Server

type Server struct {
	Addr               utils.IP4
	Connections        [2]ServerConnection
	Dispatcher         HubServerDispatcher
	Terminating        bool
	Terminated         bool
	OnError            func(any, error)
	OnAcceptConnection func(*ServerConnection, int)
	OnRejectConnection func(*net.TCPConn, int)
}

func (server *Server) ClientCount() int {
	res := 0
	for i := 0; i < len(server.Connections); i++ {
		client := &server.Connections[i]

		if client.Stats.IsOpen {
			res++
		}
	}
	return res
}

func (server *Server) Init(addr utils.IP4) {
	for cnx := range server.Connections {
		ptr := &server.Connections[cnx]
		ptr.Terminated = true
	}
	server.Addr = addr
	server.Dispatcher.Addr = addr
}

func (server *Server) SendToClients(packet Packet) {
	for i := 0; i < len(server.Connections); i++ {
		client := &server.Connections[i]

		if client.Stats.IsOpen {
			client.Write(packet)
		}
	}
}

func (server *Server) execute() {
	addr := server.Addr.ToTCPAddr()

	listener, err := net.ListenTCP("tcp4", &addr)
	if err != nil {
		log.Err(err).Msg("TCP Server Server Error")
		return
	} else {
		log.Info().Msgf("TCP Server Server Listening on: %v", addr)
	}
	defer listener.Close()

	for server.Terminating = false; !server.Terminated; {
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()

		if err != nil {
			log.Err(err).Msg("TCP Server Server Listening Error")
		} else {
			idx := server.GetFreeSlot()

			if idx != -1 {
				connection := &server.Connections[idx]
				connection.start(*conn)

				if server.OnAcceptConnection != nil {
					server.OnAcceptConnection(connection, server.ClientCount())
				}
			} else {
				if server.OnRejectConnection != nil {
					server.OnRejectConnection(conn, server.ClientCount())
				}
			}
		}
	}

	server.Terminated = true
	server.Dispatcher.Channel <- Packet{}
}

func (server *Server) Start(bindingIP utils.IP4) {
	server.Init(bindingIP)
	go server.execute()
	go server.Dispatcher.execute()
}

func (server *Server) GetFreeSlot() int {
	for i := 0; i < len(server.Connections); i++ {
		client := &server.Connections[i]

		if client.Terminated {
			return i
		}
	}
	return -1
}

func (server *Server) Stop() {
	server.Dispatcher.Stop()
	server.Terminating = true
}
