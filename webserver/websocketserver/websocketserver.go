package websocketserver

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
	"apcs_refactored/webserver/handler"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	wsHub   *WsHub
	msgNode *messenger.Node
)

func StartWebsocketServer(n *messenger.Node) {
	msgNode = n

	wsConf := config.Config.Websocket
	wsHub = newWsHub()

	handler.InitHandler()

	go wsHub.run()
	go msgNode.ListenMessages(
		func(m *messenger.Message) bool {
			log.Debugf("[%v] Received Message: %v", msgNode.Name, m)
			msg, err := json.Marshal(m)
			if err != nil {
				log.Error("Failed to marshal message")
			}
			wsHub.privateBroadcast <- msg
			// TODO - KIOSK가 메시지 수신 시그널을 보내면 true 반환
			return true
		},
	)

	r := mux.NewRouter()
	handler.Handler(r)
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(wsHub, w, r)
	})

	address := wsConf.Server.Host + ":" + strconv.Itoa(wsConf.Server.Port)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Failed to start websocket server: %v", err)
	}
}
