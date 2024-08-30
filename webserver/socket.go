package webserver

import (
	"apcs_refactored/model"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

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
	Option    interface{} `json:"option"`
	Payload   interface{} `json:"payload"`
}

// 메시지 노드 이름
const (
	INSERT_ADMIN_PWD     = "insertAdminPwd"
	UPDATE_ADMIM_PWD     = "updateAdminPwd"
	GET_ITEM_LIST        = "getItemList"
	GET_SLOT_LIST        = "getSlotList"
	GET_SLOT_TRAY_LIST   = "getSlotTrayList"
	GET_OWNER_LIST       = "getOwnerList"
	GET_OWNER_DETAIL     = "getOwnerDetail"
	INSERT_OWNER         = "insertOwner"
	UPDATE_OWNER_INFO    = "updateOwnerInfo"
	RESET_OWNER_PASSWORD = "resetOwnerPassword"
	GET_ITEM_BY_USER     = "getItemByUser"
	GET_TRAY_Buffer_Cnt  = "getTrayBufferCnt"
	GET_ITEM_CNT         = "getItemCnt"
	GET_ITEM_CNT_BY_DATE = "getItemCntByDate"
	GET_ITEM_CNT_WEEKLY  = "getItemCntWeekly"
	GET_ITEM_CNT_MONTHLY = "getItemCntMonthly"

	GET_OWNER_ADDRESS      = "getOwnerAddress"
	GET_OWNER_ADDRESS_LIST = "getOwnerAddressList"

	SENSE_TROUBLE = "senseTrouble"
)

// WebSocket 연결을 위한 주소
var (
	host        = "apms.mipllab.com"
	conn        *websocket.Conn
	connMutex   sync.Mutex
	isConnected bool
)

const secretKey = "SecretKey"

func ConnWs() {
	backoff := time.Second
	for {
		err := connectAndListen()
		if err == nil {
			break
		}
		log.Printf("Connection error: %v. Retrying in %v...", err, backoff)
		time.Sleep(backoff)
		if backoff < time.Minute {
			backoff *= 2
		}
	}
}

func connectAndListen() error {
	u := url.URL{Scheme: "wss", Host: host, Path: "/ws"}
	log.Printf("Connecting to WebSocket: %s", u.String())

	headers := http.Header{}
	clientKey := generateClientKey(secretKey)
	headers.Set("Origin", "https://apcs.com")
	headers.Set("X-Client-Key", clientKey)

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			log.Printf("HTTP Response: %d", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Response Body: %s", body)
			resp.Body.Close()
		}
		return fmt.Errorf("dial error: %v", err)
	}

	connMutex.Lock()
	conn = c
	isConnected = true
	connMutex.Unlock()

	defer func() {
		connMutex.Lock()
		conn.Close()
		isConnected = false
		connMutex.Unlock()
	}()

	if err := sendInitialConnectionMessage(); err != nil {
		return fmt.Errorf("failed to send initial message: %v", err)
	}

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	errChan := make(chan error, 1)
	go receiveMessages(c, errChan)

	for {
		select {
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			return nil
		case <-pingTicker.C:
			if err := sendPing(c); err != nil {
				return fmt.Errorf("ping failed: %v", err)
			}
		case err := <-errChan:
			return fmt.Errorf("receive error: %v", err)
		}
	}
}

func sendPing(c *websocket.Conn) error {
	connMutex.Lock()
	defer connMutex.Unlock()

	if !isConnected {
		return fmt.Errorf("connection is closed")
	}

	if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
		log.Printf("Ping 전송 실패: %v", err)
		return err
	}
	return nil
}

// var rateLimiter = time.Tick(time.Second / 10) // 초당 최대 10개의 요청으로 제한

// // 사용자 메시지 처리
// func handleUserMessage(payload string) {
// 	<-rateLimiter // 요청 제한 적용

// 	log.Println("Received user input:", payload)

// 	msg := Message{RequestId: "id", Command: "insert", Payload: payload}
// 	err := sendMsg(msg)
// 	if err != nil {
// 		log.Errorf("Failed to send message: %v", err)
// 	}
// }

// var lastInput time.Time
// var inputDebounceTime = 500 * time.Millisecond

// // 사용자 입력 처리 고루틴
// func handleUserInput(inputChan chan<- string) {
// 	for {
// 		var payload string
// 		fmt.Scanln(&payload)
// 		now := time.Now()
// 		if now.Sub(lastInput) < inputDebounceTime {
// 			continue
// 		}
// 		lastInput = now

// 		inputChan <- payload
// 	}
// }

func receiveMessages(c *websocket.Conn, errChan chan<- error) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			errChan <- fmt.Errorf("메시지 수신 에러: %v", err)
			return
		}
		reqMsg := &ReqMsg{}
		if err := json.Unmarshal(message, reqMsg); err != nil {
			log.Printf("JSON 파싱 에러: %v", err)
			continue
		}
		log.Printf("서버로부터 메시지 수신: %s\n", reqMsg)
		go handleMessage(reqMsg)
	}
}

func sendInitialConnectionMessage() error {
	u := uuid.New()
	name, err := model.SelectIbName()
	if err != nil {
		return fmt.Errorf("failed to select IB name: %v", err)
	}
	msg := Message{RequestId: u.String(), Command: "conn", Payload: name}
	return sendMsg(msg)
}

func handleMessage(reqMsg *ReqMsg) {
	var res Message

	switch reqMsg.Command {
	case INSERT_ADMIN_PWD:
		res = insertAdminPwd(reqMsg)
	case UPDATE_ADMIM_PWD:
		res = updateAdminPwd(reqMsg)
	case GET_ITEM_LIST:
		res = getItemList(reqMsg)
	case GET_SLOT_LIST:
		res = getSlotList(reqMsg)
	case GET_SLOT_TRAY_LIST:
		res = getSlotTrayList(reqMsg)
	case GET_OWNER_LIST:
		res = getOwnerList(reqMsg)
	case GET_OWNER_DETAIL:
		res = getOwnerDetail(reqMsg)
	case INSERT_OWNER:
		res = insertOwner(reqMsg)
	case UPDATE_OWNER_INFO:
		res = updateOwnerInfo(reqMsg)
	case GET_ITEM_BY_USER:
		res = getItemByUser(reqMsg)
	case GET_TRAY_Buffer_Cnt:
		res = getTrayBufferCnt(reqMsg)
	case GET_OWNER_ADDRESS:
		res = getOwnerAddress(reqMsg)
	case GET_OWNER_ADDRESS_LIST:
		res = getOwnerAddressList(reqMsg)
	case RESET_OWNER_PASSWORD:
		res = resetOwnerPassword(reqMsg)
	case GET_ITEM_CNT:
		res = getItemCnt(reqMsg)
	case GET_ITEM_CNT_BY_DATE:
		res = getItemCntByDate(reqMsg)
	case GET_ITEM_CNT_WEEKLY:
		res = getItemCntWeekly(reqMsg)
	case GET_ITEM_CNT_MONTHLY:
		res = getItemCntMonthly(reqMsg)
	default:
		log.Printf("Unknown command: %s", reqMsg.Command)
		return
	}

	sendMsg(res)
}

func generateClientKey(secretKey string) string {
	// HMAC-256 해시 함수 사용하여 클라이언트 키 생성
	h := hmac.New(sha256.New, []byte(secretKey))
	return hex.EncodeToString(h.Sum(nil))
}

func sendMsg(msg Message) error {
	connMutex.Lock()
	defer connMutex.Unlock()

	if !isConnected {
		return fmt.Errorf("connection is closed")
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("JSON 인코딩 에러: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		return fmt.Errorf("메시지 전송 에러: %v", err)
	}
	return nil
}

func getOwnerAddress(data *ReqMsg) Message {
	id := data.Payload

	address, err := model.SelectAddressByOwnerId(id)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	msg := &Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: address}
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
		msg.Status = "FAIL"
		msg.Payload = "master비밀번호가 이미 존재합니다"
	} else {
		_, err := model.InsertAdminPwd(password)
		if err != nil {
			log.Error(err)
			msg.RequestId = data.RequestId
			msg.Command = data.Command
			msg.Status = "FAIL"
			msg.Payload = err.Error()
		}
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "OK"
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
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	msg.RequestId = data.RequestId
	msg.Command = data.Command
	msg.Status = "OK"
	msg.Payload = "수정완료"

	log.Println("sendToServer: ", msg)
	return *msg
}

type ItemStatus struct {
	Input  bool `json:"input"`
	Output bool `json:"output"`
}

func getItemList(data *ReqMsg) Message {

	option, _ := json.Marshal(data.Option)
	payload, _ := json.Marshal(data.Payload)

	itemOption := &model.ItemOption{}

	erro := json.Unmarshal(payload, itemOption)
	if erro != nil {
		log.Error(erro)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: erro.Error()}
		return msg
	}

	log.Println("itemOption: ", itemOption)

	itemStatus := &ItemStatus{}

	erro = json.Unmarshal(option, itemStatus)
	if erro != nil {
		log.Error(erro)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: erro.Error()}
		return msg
	}

	log.Println("option: ", itemStatus)

	msg := &Message{}
	var itemList []model.ItemListResponse
	var err error

	if !itemStatus.Input && itemStatus.Output {
		log.Println("SelectOutputItemList")

		itemList, err = model.SelectOutputItemList(itemOption)

	} else if itemStatus.Input && !itemStatus.Output {
		log.Println("SelectStoreItemList")

		itemList, err = model.SelectStoreItemList(itemOption)
	} else {
		log.Println("SelectItemList")

		itemList, err = model.SelectItemList(itemOption)
	}
	// switch option {
	// case "input":
	// 	itemList, err = model.SelectInputItemList(itemOption)
	// case "output":
	// 	itemList, err = model.SelectOutputItemList(itemOption)
	// case "store":
	// 	itemList, err = model.SelectStoreItemList(itemOption)
	// }

	if err != nil {
		log.Error(err)
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "FAIL"
		msg.Payload = err.Error()
	}
	msg.RequestId = data.RequestId
	msg.Command = data.Command
	msg.Status = "OK"
	msg.Payload = itemList
	log.Println("sendToServer: ", msg)
	return *msg
}

func getSlotList(data *ReqMsg) Message {
	slotList, err := model.SelectSlotList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: slotList}
	log.Println("sendToServer: ", msg)
	return msg
}

func getSlotTrayList(data *ReqMsg) Message {
	trayList, err := model.SelectTrayList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: trayList}
	log.Println("sendToServer: ", msg)
	return msg
}

func getOwnerList(data *ReqMsg) Message {
	owner, err := model.SelectOwnerList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: owner}
	log.Println("sendToServer: ", msg)
	return msg
}

func getOwnerDetail(data *ReqMsg) Message {
	owner, err := model.SelectOwnerDetail(data.Option)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: owner}
	log.Println("sendToServer: ", msg)
	return msg
}

type Owner struct {
	OwnerId  int    `json:"owner_id"`
	Nm       string `json:"nm"`
	Address  string `json:"address"`
	Password string `json:"password"`
	PhoneNum string `json:"phone_num"`
}

func insertOwner(data *ReqMsg) Message {
	payload, _ := json.Marshal(data.Payload)
	owner := &Owner{}
	err := json.Unmarshal(payload, owner)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	log.Println("owner: ", owner)

	rslt, _ := model.SelectOwnerIdByAddress(owner.Address)
	log.Println("owner: ", rslt)

	if rslt != 0 {
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: "해당 유저가 이미 존재합니다"}
		log.Error("해당 유저가 이미 존재합니다")
		return msg
	} else {
		ownerCreateRequest := model.OwnerCreateRequest{Nm: owner.Nm, PhoneNum: owner.PhoneNum, Address: owner.Address, Password: owner.Password}
		_, err := model.InsertOwner(ownerCreateRequest)
		if err != nil {
			msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
			log.Error(err)
			return msg
		}
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: "등록완료"}
		return msg
	}
}

func updateOwnerInfo(data *ReqMsg) Message {
	payload, _ := json.Marshal(data.Payload)

	owner := &Owner{}
	err := json.Unmarshal(payload, owner)
	log.Println("=====: ", owner)

	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	log.Println("owner: ", owner)

	// rslt, _ := model.SelectOwnerIdByAddress(owner.Address)
	// log.Println("owner: ", rslt)

	// if rslt != 0 {
	// 	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: "해당 유저가 이미 존재합니다"}
	// 	log.Error("해당 유저가 이미 존재합니다")
	// 	return msg

	// } else {
	_, err = model.UpdateOwnerInfo(model.OwnerUpdateRequest{Nm: owner.Nm, PhoneNum: owner.PhoneNum, Address: owner.Address, OwnerId: int64(owner.OwnerId)})

	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}

		return msg
	}

	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: "수정완료"}
	log.Println("sendToServer: ", msg)

	return msg
	//}
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
		msg.Status = "FAIL"
		msg.Payload = err.Error()
	}
	msg.RequestId = data.RequestId
	msg.Command = data.Command
	msg.Status = "OK"
	msg.Payload = itemList
	log.Println("sendToServer: ", msg)
	return *msg
}

type TrayInfo struct {
	TrayCount int         `json:"trayCnt"`
	List      interface{} `json:"list"`
}

func getTrayBufferCnt(data *ReqMsg) Message {
	//list := trayBuffer.Buffer.Get()
	// TODO - 스택에 있는 개수와 DB개수 비교
	tray_buffer, _ := model.SelectTrayBufferState()
	//payload := &TrayInfo{TrayCount: tray_buffer.Count, List: list}
	payload := &TrayInfo{TrayCount: tray_buffer.Count}
	log.Println(payload)

	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: payload}
	log.Println("sendToServer: ", msg)
	return msg
}

func getOwnerAddressList(data *ReqMsg) Message {
	owner, err := model.SelectOwnerAddressList()
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: owner}
	log.Println("sendToServer: ", msg)
	return msg
}

func resetOwnerPassword(data *ReqMsg) Message {
	payload, _ := json.Marshal(data.Payload)

	owner := &Owner{}
	err := json.Unmarshal(payload, owner)
	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}
		return msg
	}
	log.Println("owner: ", owner)

	_, err = model.ResetOwnerPassword(model.OwnerPwdRequest{Password: owner.Password, OwnerId: owner.OwnerId})

	if err != nil {
		log.Error(err)
		msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: err.Error()}

		return msg
	}

	msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "OK", Payload: "수정완료"}
	log.Println("sendToServer: ", msg)

	return msg
}

func getItemCnt(data *ReqMsg) Message {
	itemDate := &model.ItemCntReq{}

	if data.Option != "" {
		option, _ := json.Marshal(data.Option)

		erro := json.Unmarshal(option, itemDate)
		if erro != nil {
			log.Error(erro)
			msg := Message{RequestId: data.RequestId, Command: data.Command, Status: "FAIL", Payload: erro.Error()}
			return msg
		}
		log.Println("itemDate: ", itemDate)
	}

	msg := &Message{}

	itemCnt, err := model.SelectItemCnt(itemDate)

	if err != nil {
		log.Error(err)
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "FAIL"
		msg.Payload = err.Error()
	}

	msg.RequestId = data.RequestId
	msg.Command = data.Command
	msg.Status = "OK"
	msg.Payload = itemCnt
	log.Println("sendToServer: ", msg)
	return *msg
}

func SendEvent(data string) {
	event := Message{RequestId: "0", Command: SENSE_TROUBLE, Status: "OK", Payload: data}
	log.Println("sendEventToServer: ", event)

	// JSON 인코딩
	jsonMsg, err := json.Marshal(event)
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

func getItemCntByDate(data *ReqMsg) Message {

	msg := &Message{}

	itemCntList, err := model.SelectItemCntByDate()

	if err != nil {
		log.Error(err)
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "FAIL"
		msg.Payload = err.Error()
	} else {
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "OK"
		msg.Payload = itemCntList
	}

	log.Println("sendToServer: ", msg)
	return *msg
}
func getItemCntWeekly(data *ReqMsg) Message {

	msg := &Message{}

	itemCntList, err := model.SelectItemCntWeekly()

	if err != nil {
		log.Error(err)
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "FAIL"
		msg.Payload = err.Error()
	} else {
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "OK"
		msg.Payload = itemCntList
	}

	log.Println("sendToServer: ", msg)
	return *msg
}
func getItemCntMonthly(data *ReqMsg) Message {

	msg := &Message{}

	itemCntList, err := model.SelectItemCntMonthly()

	if err != nil {
		log.Error(err)
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "FAIL"
		msg.Payload = err.Error()
	} else {
		msg.RequestId = data.RequestId
		msg.Command = data.Command
		msg.Status = "OK"
		msg.Payload = itemCntList
	}

	log.Println("sendToServer: ", msg)
	return *msg
}
