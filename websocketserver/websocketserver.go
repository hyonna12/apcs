package websocketserver

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
	"apcs_refactored/websocketserver/handler"
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

	wsConf := config.Conf.Websocket
	wsHub = newWsHub()

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

	address := wsConf.Server.Host + ":" + strconv.Itoa(wsConf.Server.Port)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Failed to start websocket server: %v", err)
	}
}
