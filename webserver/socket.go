package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc/trayBuffer"
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
	GET_SLOT_TRAY_LIST  = "getSlotTrayList"
	GET_OWNER_LIST      = "getOwnerList"
	INSERT_OWNER        = "insertOwner"
	UPDATE_OWNER_INFO   = "updateOwnerInfo"
	GET_ITEM_BY_USER    = "getItemByUser"
	GET_TRAY_Buffer_Cnt = "getTrayBufferCnt"

	GET_OWNER_ADDRESS      = "getOwnerAddress"
	GET_OWNER_ADDRESS_LIST = "getOwnerAddressList"
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
				res := insertAdminPwd(reqMsg)
				sendMsg(res)
			case UPDATE_ADMIM_PWD:
				res := updateAdminPwd(reqMsg)
				sendMsg(res)
			case GET_ITEM_LIST:
				res := getItemList(reqMsg)
				sendMsg(res)

			case GET_SLOT_LIST:
				res := getSlotList(reqMsg)
				sendMsg(res)

			case GET_SLOT_TRAY_LIST:
				res := getSlotTrayList(reqMsg)
				sendMsg(res)
			case GET_OWNER_LIST:
				res := getOwnerList(reqMsg)
				sendMsg(res)
			case INSERT_OWNER:
				res := insertOwner(reqMsg)
				sendMsg(res)
			case UPDATE_OWNER_INFO:
				res := updateOwnerInfo(reqMsg)
				sendMsg(res)
			case GET_ITEM_BY_USER:
				res := getItemByUser(reqMsg)
				sendMsg(res)
			case GET_TRAY_Buffer_Cnt:
				res := getTrayBufferCnt(reqMsg)
				sendMsg(res)
			case GET_OWNER_ADDRESS:
				res := getOwnerAddress(reqMsg)
				sendMsg(res)
			case GET_OWNER_ADDRESS_LIST:
				res := getOwnerAddressList(reqMsg)
				sendMsg(res)

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

func getOwnerAddress(data *ReqMsg) Message {
	id := data.Payload

	address, err := model.SelectAddressByOwnerId(id)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	msg := &Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: address}
	log.Println("sendToServer: ", msg)
	return *msg
}

func insertAdminPwd(data *ReqMsg) Message {
	password := data.Payload

	// result, _ := json.Marshal(payload)
	// log.Println(result)

	// err := mapstructure.Decode(payload, &admin)
	// if err != nil {
	// 	log.Error(err)
	// }
	// log.Println(admin)
	//h_pwd := sha256.Sum256([]byte(password.(string)))
	bool, _ := model.SelectExistPassword()
	msg := &Message{}
	if bool == 1 {
		log.Println("master 비밀번호가 존재합니다")
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "fail"
		msg.Payload = "master비밀번호가 이미 존재합니다"
	} else {
		_, err := model.InsertAdminPwd(password)
		if err != nil {
			log.Error(err)
			msg.RequestId = data.RequestId
			msg.Command = data.Command
			msg.Status = "fail"
			msg.Payload = err.Error()
		}
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "fail"
		msg.Payload = "등록 완료"
		log.Println("sendToServer: ", msg)
	}
	return *msg
}

func updateAdminPwd(data *ReqMsg) Message {
	password := data.Payload

	msg := &Message{}
	_, err := model.InsertAdminPwd(password)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	msg.RequestId = data.RequestId
	msg.Command = data.Command
	msg.Status = "ok"
	msg.Payload = "수정완료"

	log.Println("sendToServer: ", msg)
	return *msg
}

func getItemList(data *ReqMsg) Message {
	option := data.Option
	msg := &Message{}
	var itemList []model.ItemListResponse
	var err error

	switch option {
	case "input":
		itemList, err = model.SelectInputItemList()
	case "output":
		itemList, err = model.SelectOutputItemList()
	case "store":
		itemList, err = model.SelectStoreItemList()
	}
	if err != nil {
		log.Error(err)
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "fail"
		msg.Payload = err.Error()
	}
	msg.RequestId = data.RequestId
	msg.Command = data.Command
	msg.Status = "ok"
	msg.Payload = itemList
	log.Println("sendToServer: ", msg)
	return *msg
}

func getSlotList(data *ReqMsg) Message {
	slotList, err := model.SelectSlotList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: slotList}
	log.Println("sendToServer: ", msg)
	return msg
}

func getSlotTrayList(data *ReqMsg) Message {
	trayList, err := model.SelectTrayList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: trayList}
	log.Println("sendToServer: ", msg)
	return msg
}

func getOwnerList(data *ReqMsg) Message {
	owner, err := model.SelectOwnerList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: owner}
	log.Println("sendToServer: ", msg)
	return msg
}

type Owner struct {
	OwnerId  int    `json:"id"`
	Address  string `json:"address"`
	Password string `json:"password"`
	PhoneNum string `json:"phoneNum"`
}

func insertOwner(data *ReqMsg) Message {
	payload, _ := json.Marshal(data.Payload)
	owner := &Owner{}
	err := json.Unmarshal(payload, owner)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	log.Println("owner: ", owner)

	bool, _ := model.SelectOwnerIdByAddress(owner.Address)
	if bool == 1 {
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: "해당 유저가 이미 존재합니다"}
		log.Error("해당 유저가 이미 존재합니다")
		return msg
	} else {
		ownerCreateRequest := model.OwnerCreateRequest{PhoneNum: owner.PhoneNum, Address: owner.Address, Password: owner.Password}
		_, err := model.InsertOwner(ownerCreateRequest)
		if err != nil {
			msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
			log.Error(err)
			return msg
		}
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: "등록완료"}
		return msg
	}
}

func updateOwnerInfo(data *ReqMsg) Message {
	payload, _ := json.Marshal(data.Payload)
	owner := &Owner{}
	err := json.Unmarshal(payload, owner)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	log.Println("owner: ", owner)

	// TODO - 해당 id를 가진 owner 있는지 확인
	_, err = model.UpdateOwnerInfo(model.OwnerUpdateRequest{OwnerId: int64(owner.OwnerId), PhoneNum: owner.PhoneNum, Password: owner.Password})
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}

	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: "수정완료"}
	log.Println("sendToServer: ", msg)
	return msg
}

func getItemByUser(data *ReqMsg) Message {
	option := data.Option
	msg := &Message{}
	var itemList []model.ItemListResponse
	var err error
	owner_id := data.Payload
	switch option {
	case "input":
		itemList, err = model.SelectInputItemByUser(owner_id)
	case "output":
		itemList, err = model.SelectOutputItemByUser(owner_id)
	case "store":
		itemList, err = model.SelectStoreItemByUser(owner_id)
	}
	if err != nil {
		log.Error(err)
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "fail"
		msg.Payload = err.Error()
	}
	msg.RequestId = data.RequestId
	msg.Command = data.Command
	msg.Status = "ok"
	msg.Payload = itemList
	log.Println("sendToServer: ", msg)
	return *msg
}

type TrayInfo struct {
	TrayCount int         `json:"trayCnt"`
	List      interface{} `json:"list"`
}

func getTrayBufferCnt(data *ReqMsg) Message {
	list := trayBuffer.Buffer.Get()
	tray_buffer, _ := model.SelectTrayBufferState()
	payload := &TrayInfo{TrayCount: tray_buffer.Count, List: list}
	log.Println(payload)
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: payload}
	log.Println("sendToServer: ", msg)
	return msg
}

func getOwnerAddressList(data *ReqMsg) Message {
	owner, err := model.SelectOwnerAddressList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "fail", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "ok", Payload: owner}
	log.Println("sendToServer: ", msg)
	return msg
}
