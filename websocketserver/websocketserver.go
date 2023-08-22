package websocketserver

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
	r.HandleFunc("/", Home)
	/* input */
	r.HandleFunc("/input/regist_delivery", RegistDelivery)
	r.HandleFunc("/input/input_item", InputItem)
	r.HandleFunc("/input/input_item_error", InputItemError)
	r.HandleFunc("/input/regist_owner", RegistOwner)
	r.HandleFunc("/input/regist_owner_error", RegistOwnerError)
	r.HandleFunc("/input/complete_input_item", CompleteInputItem)
	r.HandleFunc("/input/cancel_input_item", CancelInputItem)
	/* output */
	r.HandleFunc("/output/regist_address", RegistAddress)
	r.HandleFunc("/output/regist_address_error", RegistAddressError)
	r.HandleFunc("/output/item_list", ItemList)
	r.HandleFunc("/output/item_list_error", ItemListError)
	r.HandleFunc("/output/complete_output_item", CompleteOutputItem)

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsHub, w, r)
	})

	http.Handle("/", r)
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("/static/")))
	r.Handle("/static/", staticHandler)

	address := wsConf.Server.Host + ":" + strconv.Itoa(wsConf.Server.Port)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Failed to start websocket server: %v", err)
	}
}
