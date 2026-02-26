package web

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type SocketClient struct {
	Subscriptions uint64
	service       *SocketService
	conn          *websocket.Conn
	send          chan *SocketPayload
}

func (s *SocketClient) readSocket() {
	defer func() {
		s.service.unregister <- s
		s.conn.Close()
	}()

	s.conn.SetReadLimit(s.service.MaxReadSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(s.service.PongEvery))
	s.conn.SetPongHandler(func(string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.service.PongEvery))
		return nil
	})

	for {
		_, message, err := s.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debug().Msgf("SocketClient.readSocket - %v", err)
			}
			break
		}

		s.handleMessage(message)
	}
}

func (s *SocketClient) writeSocket() {
	ticker := time.NewTicker(s.service.PingEvery)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()

	for {
		select {
		case message, ok := <-s.send:
			if ok {
				if message.Subscription == 0 || message.Subscription&s.Subscriptions != 0 {
					_ = s.conn.SetWriteDeadline(time.Now().Add(s.service.WriteDeadline))
					if !ok {
						_ = s.conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}

					w, err := s.conn.NextWriter(websocket.BinaryMessage)
					if err != nil {
						return
					}

					if _, err = w.Write(message.Payload); err != nil {
						return
					}

					if err = w.Close(); err != nil {
						return
					}
				}
			}

		case <-ticker.C:
			_ = s.conn.SetWriteDeadline(time.Now().Add(s.service.WriteDeadline))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *SocketClient) handleMessage(message []byte) {
	requestMsg := SocketMessage{}
	if err := requestMsg.LoadBytes(message); err != nil {
		log.Err(err).Msg("SocketClient.handleMessage")
		// TODO: Consider responding to the error
		return
	}

	//var jsonData map[string]interface{}
	//
	//if err := json.Unmarshal(message, &jsonData); err != nil {
	//	var errMap = make(map[string]interface{})
	//	errMap["type"] = "Error"
	//	errMap["error"] = err.Error()
	//	payload, _ := json.Marshal(errMap)
	//
	//	sm := &SocketPayload{
	//		Subscription: 0,
	//		Payload:      payload,
	//	}
	//	s.send <- sm
	//	return
	//}

	switch requestMsg.GetType("") {
	case "my-subscriptions-request":
		responseMsg := SocketMessage{}
		responseMsg.Init()
		responseMsg.SetType("my-subscriptions-response")
		responseMsg.SetInt("Value", int(s.Subscriptions))
		s.send <- responseMsg.ToPayload(0)
	}
}
