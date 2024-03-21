package webserver

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// 메시지 구조체
type Message struct {
	RequestId string      `json:"request_id"`
	Command   string      `json:"command"`
	Status    string      `json:"status"`
	Payload   interface{} `json:"payload"`
}

// WebSocket 연결을 위한 주소
var url = "ws://localhost:8080"

// var id string
var conn *websocket.Conn

func ConnWs() {
	// WebSocket 연결
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("연결 실패:", err)
	} else {
		conn = c
		// 택배함 id 값 전송
		u := uuid.New()
		//id, _ := model.SelectIbId()
		msg := Message{RequestId: u.String(), Command: "conn", Payload: 2}
		sendMsg(msg)
	}
	defer c.Close()

	// 비동기적인 메시지 수신을 위한 고루틴 실행
	go func() {
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("메시지 수신 에러:", err)
				return
			}
			fmt.Printf("서버로부터 메시지 수신: %s\n", message)
		}
	}()

	// signal 받을 채널 생성 및 등록
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		// signal 받으면 프로세스 종료
		case <-interrupt:
			fmt.Println("CTRL+C가 입력되어 종료합니다.")
			return
		// 사용자 입력을 받아 서버에 메시지 전송
		default:
			var content string
			fmt.Print("전송할 메시지를 입력하세요: ")
			fmt.Scanln(&content)

			// 메시지 생성
			msg := Message{RequestId: "id", Command: "send", Payload: content}

			sendMsg(msg)
		}
	}
}

func sendMsg(msg Message) {
	// JSON 인코딩
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Println("JSON 인코딩 에러:", err)
		return
	}

	// 메시지 전송
	err = conn.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		log.Println("메시지 전송 에러:", err)
		return
	}
}
