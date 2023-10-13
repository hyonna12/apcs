package webserver

import (
	log "github.com/sirupsen/logrus"
)

type WsHub struct {
	clients          map[*Client]bool
	publicBroadcast  chan []byte
	privateBroadcast chan []byte
	register         chan *Client
	unregister       chan *Client
}

func newWsHub() *WsHub {
	return &WsHub{
		clients:          make(map[*Client]bool),
		publicBroadcast:  make(chan []byte),
		privateBroadcast: make(chan []byte),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
	}
}

func (h *WsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		// 웹소켓 클라이언트로부터 들어온 메시지는 메신저 포함해서 전파
		case message := <-h.publicBroadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

			//메시지를 래핑해서 메신저에 전달
			err := msgNode.SpreadMessage(message)
			if err != nil {
				// TODO - 메시지 응답 없음 관련 에러 처리
				log.Error(err)
			}
		// 외부에서 메시지 노드로 들어온 메시지는 메신저 허브로 전파하지 않음
		case message := <-h.privateBroadcast:
			log.Debugf("Broadcast message from public to private: %v", string(message))
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
