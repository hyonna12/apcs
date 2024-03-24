package webserver

import (
	"apcs_refactored/model"
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
	RequestId string      `json:"requestId"`
	Command   string      `json:"command"`
	Status    string      `json:"status"`
	Payload   interface{} `json:"payload"`
}

type ReqMsg struct {
	RequestId string      `json:"requestId"`
	Command   string      `json:"command"`
	Option    string      `json:"option"`
	Payload   interface{} `json:"payload"`
}

// 메시지 노드 이름
const (
	INSERT_ADMIN_PWD    = "insertAdminPwd"
	UPDATE_ADMIM_PWD    = "updateAdminPwd"
	GET_ITEM_LIST       = "getItemList"
	GET_SLOT_LIST       = "getSlotList"
	GET_TRAY_LIST       = "getTrayList"
	GET_OWNER_LIST      = "getOwnerList"
	INSERT_OWNER        = "insertOwner"
	UPDATE_OWNER_INFO   = "updateOwnerInfo"
	GET_ITEM_BY_USER    = "getItemByUser"
	GET_TRAY_Buffer_Cnt = "getTrayBufferCnt"
)

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
		id, _ := model.SelectIbId()
		msg := Message{RequestId: u.String(), Command: "conn", Payload: id}
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

			reqMsg := &ReqMsg{}
			json.Unmarshal(message, reqMsg)
			log.Printf("서버로부터 메시지 수신: %s\n", reqMsg)

			switch reqMsg.Command {
			case INSERT_ADMIN_PWD:
				insertAdminPwd()
			case UPDATE_ADMIM_PWD:
				updateAdminPwd()
			case GET_ITEM_LIST:
				getItemList()
			case GET_SLOT_LIST:
				getSlotList()
			case GET_TRAY_LIST:
				getTrayList()
			case GET_OWNER_LIST:
				getOwnerList()
			case INSERT_OWNER:
				insert_owner()
			case UPDATE_OWNER_INFO:
				updateOwnerInfo()
			case GET_ITEM_BY_USER:
				getItemByUser()
			case GET_TRAY_Buffer_Cnt:
				getTrayBufferCnt()
			case "getOwnerAddress":
				getOwnerAddress(reqMsg)
			}
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
			var payload string
			//fmt.Print("전송할 메시지를 입력하세요: ")
			fmt.Scanln(&payload)

			// 메시지 생성
			msg := Message{RequestId: "id", Command: "insert", Payload: payload}

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

func getOwnerAddress(data *ReqMsg) {
	id := data.Payload
	address, _ := model.SelectAddressByOwnerId(id)
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: address}
	log.Println("sendToServer: ", msg)
	sendMsg(msg)
}

func insertAdminPwd() {
	log.Println("insertAdminPwd")
}
func updateAdminPwd() {
	log.Println("updateAdminPwd")
}
func getItemList() {
	log.Println("getItemList")
}
func getSlotList() {
	log.Println("getSlotList")
}
func getTrayList() {
	log.Println("getTrayList")
}
func getOwnerList() {
	log.Println("getOwnerList")
}
func insert_owner() {
	log.Println("insert_owner")
}
func updateOwnerInfo() {
	log.Println("updateOwnerInfo")

}
func getItemByUser() {
	log.Println("getItemByUser")

}
func getTrayBufferCnt() {
	log.Println("getTrayBufferCnt")
}
