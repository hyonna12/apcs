package websocketserver

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

var (
	wsHub   *WsHub
	msgNode *messenger.Node
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	http.ServeFile(w, r, "websocketserver/kiosk.html")
}

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

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsHub, w, r)
	})
	http.HandleFunc("/", serveHome)

	address := wsConf.Server.Host + ":" + strconv.Itoa(wsConf.Server.Port)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Failed to start websocket server: %v", err)
	}
}
