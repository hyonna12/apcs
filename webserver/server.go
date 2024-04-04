package webserver

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
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

func StartWebserver(n *messenger.Node) {
	msgNode = n

	webConfig := config.Config.Webserver
	wsHub = newWsHub()

	requestList = make(map[int64]*request)

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

	// 템플릿 엔진 초기화
	initTemplate()

	r := mux.NewRouter()
	Handler(r)
	http.Handle("/", r)

	address := webConfig.Server.Host + ":" + strconv.Itoa(webConfig.Server.Port)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Failed to start websocket server: %v", err)
	}
}

func broadcastToPrivate(message []byte) {
	wsHub.privateBroadcast <- message
}
