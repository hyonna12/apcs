package websocketserver

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// writeWait - peer에게 메시지를 쓸 때 허용되는 시간
	writeWait = 10 * time.Second

	// pongWait - peer로부터 온 pong 메시지를 읽을 때 허용되는 시간
	pongWait = 60 * time.Second

	// pingPeriod - 이 시간 간격마다 peer에게 ping 메시지 송부. pong 간격보다 작아야 함
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize - peer로부터 받는 메시지 최대 크기 (byte)
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	wsHub *WsHub
	conn  *websocket.Conn
	// outbound(브라우저로 나가는) 메시지를 위한 buffered channel
	send chan []byte
}

// readPump 커넥션에서 들어오는 메시지를 허브로 전달
func (c *Client) readPump() {
	defer func() {
		c.wsHub.register <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		// c.conn.ReadMessage()는 웹소켓 연결이 들어올 때까지 block되기 때문에
		// busy waiting이 아님
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.wsHub.publicBroadcast <- message
	}
}

// writePump 허브에서 들어온 메시지를 커넥션으로 전달
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *WsHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	client := &Client{wsHub: hub, conn: conn, send: make(chan []byte, 256)}
	client.wsHub.register <- client

	go client.readPump()
	go client.writePump()
}
