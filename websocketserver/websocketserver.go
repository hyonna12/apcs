package websocketserver

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	wsHub   *WsHub
	msgNode *messenger.Node
)

var templ = func() *template.Template {
	t := template.New("")
	err := filepath.Walk("websocketserver/views/", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			fmt.Println(path)
			_, err = t.ParseFiles(path)
			if err != nil {
				fmt.Println(err)
			}
		}
		return err
	})

	if err != nil {
		panic(err)
	}
	return t
}()

type Page struct {
	Title string
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "main", &Page{Title: "Home"})

	//http.ServeFile(w, r, "websocketserver/views/main.html")
}

/*
	 func serveHome(w http.ResponseWriter, r *http.Request) {
		// /p := path.Dir("websocketserver/views/index.html")
		// set header
		w.Header().Set("Content-type", "text/html")
		http.ServeFile(w, r, "websocketserver/views/main.html")
	}
*/
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
	r.HandleFunc("/", serveHome)
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
