package conn

import (
	"apcs_refactored/webserver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/future-architect/go-mcprotocol/mcp"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Task struct {
	TaskName string
	Address  string
}

var TaskList = map[string]Task{
	"Door":           {"Door", "D100"},  // (0: closed, 1: open)
	"Light":          {"Light", "D101"}, // (0: off, 1: on)
	"SetTemperature": {"SetTemperature", "D102"},
	"DetectFire":     {"DetectFire", "D103"}, // (0: no fire, 1: fire detected)
	"Weight":         {"Weight", "D104"},
	"Height":         {"Height", "D105"},
}

// PLC 서버와의 웹소켓 연결을 위한 변수들
var (
	plcConn        *websocket.Conn
	isPlcConnected bool
)

// PLC 메시지 구조체
type PlcMessage struct {
	Type     string      `json:"type"`
	Id       string      `json:"id"`
	TargetId string      `json:"targetId,omitempty"`
	Content  interface{} `json:"content,omitempty"`
}

// PLC 서버 연결 함수
func ConnectPlcServer() {
	for {
		conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:6000", nil)
		if err != nil {
			log.Errorf("Failed to connect to PLC server: %v", err)
			time.Sleep(5 * time.Second) // 재연결 전 대기
			continue
		}

		plcConn = conn
		isPlcConnected = true
		log.Info("Connected to PLC server")

		// 초기 연결 메시지 전송
		initMsg := PlcMessage{
			Type: "conn",
			Id:   "apcs_client",
		}
		err = sendPlcMessage(initMsg)
		if err != nil {
			log.Error("Failed to send initial message:", err)
			plcConn.Close()
			continue
		}

		// 메시지 수신 처리
		go handlePlcMessages()

		break
	}
}

// PLC 메시지 전송 함수
func sendPlcMessage(msg PlcMessage) error {
	if !isPlcConnected {
		return fmt.Errorf("PLC server not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = plcConn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		isPlcConnected = false
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

// PLC 메시지 수신 처리 함수
func handlePlcMessages() {
	for {
		_, message, err := plcConn.ReadMessage()
		if err != nil {
			log.Errorf("Error reading message: %v", err)
			isPlcConnected = false
			plcConn.Close()
			go ConnectPlcServer() // 재연결 시도
			return
		}

		// 서버로부터 받은 메시지 출력
		msgStr := string(message)
		log.Tracef("Raw message from PLC server: %s", msgStr)

		// 초기 연결 메시지는 무시
		if strings.Contains(msgStr, "connect to server.js") {
			continue
		}

		// "Message from server: 화재" 형식의 메시지 처리
		if strings.Contains(msgStr, "Message from server:") {
			// 메시지에서 실제 내용 추출
			content := strings.TrimPrefix(msgStr, "Message from server: ")
			content = strings.TrimSpace(content)

			log.Debugf("Extracted trouble content: %s", content)

			// 트러블 처리
			SenseTrouble(content)
			continue
		}

		// JSON 메시지 처리 (필요한 경우)
		var plcMsg PlcMessage
		if err := json.Unmarshal(message, &plcMsg); err == nil {
			if plcMsg.Type == "trouble" {
				troubleType, ok := plcMsg.Content.(string)
				if ok {
					SenseTrouble(troubleType)
				}
			}
		}
	}
}

// 문자열에 특정 키워드가 포함되어 있는지 확인하는 헬퍼 함수
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// PLC 서버로 트러블 메시지 전송
func SendTroubleToPlc(troubleType string) {
	if !isPlcConnected {
		log.Error("PLC server not connected")
		return
	}

	msg := PlcMessage{
		Type:    "trouble",
		Id:      "apcs_client",
		Content: troubleType,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("Failed to marshal trouble message: %v", err)
		return
	}

	err = plcConn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Errorf("Failed to send trouble message: %v", err)
		isPlcConnected = false
		return
	}

	log.Debugf("Sent trouble message to PLC server: %s", troubleType)
}

// 트러블 감지 함수 수정
func SenseTrouble(data string) {
	log.Debug(data)

	switch data {
	case "화재":
		log.Infof("[PLC] 화재발생")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error("Error changing view:", err)
		}
		webserver.SendEvent("화재")

	case "물품 끼임":
		log.Infof("[PLC] 물품 끼임")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		webserver.SendEvent("물품 끼임")

	case "물품 낙하":
		log.Infof("[PLC] 물품 낙하")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		webserver.SendEvent("물품 낙하")

	case "이물질 감지":
		log.Infof("[PLC] 이물질 감지")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		webserver.SendEvent("이물질 감지")
	}

	// PLC 서버에 트러블 상태 전송
	if isPlcConnected {
		msg := PlcMessage{
			Type:    "trouble_ack",
			Id:      "apcs_client",
			Content: data,
		}
		if err := sendPlcMessage(msg); err != nil {
			log.Error("Failed to send trouble acknowledgment:", err)
		}
	}
}

var McClient mcp.Client

// InitConnPlc 함수 수정
func InitConnPlc() {
	// PLC 서버 연결 시작
	go ConnectPlcServer()
}

func ParseAddress(address string) (string, int64, error) {
	deviceName := string(address[0])
	offset, err := strconv.ParseInt(address[1:], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid address format: %s", address)
	}
	return deviceName, offset, nil
}

func read(taskName string) error {
	task, exists := TaskList[taskName]
	if !exists {
		return fmt.Errorf("task %s not found", taskName)
	}
	fmt.Printf("Start Task: %s\n", taskName)

	deviceName, offset, err := ParseAddress(task.Address)
	if err != nil {
		return err
	}

	response, err := McClient.Read(deviceName, offset, 1)
	if err != nil {
		return fmt.Errorf("error reading from address %s: %v", task.Address, err)
	}

	if len(response) < 2 {
		return fmt.Errorf("invalid response length")
	}

	value := binary.LittleEndian.Uint16(response[11:13])
	//value := binary.LittleEndian.Uint16(response)
	fmt.Printf("Read value from address %s[%s]: %v\n", taskName, task.Address, value)
	return nil
}

func write(taskName string, value uint16) error {
	task, exists := TaskList[taskName]
	if !exists {
		return fmt.Errorf("task %s not found", taskName)
	}

	fmt.Printf("Start Task: %s\n", task.TaskName)

	deviceName, offset, err := ParseAddress(task.Address)
	if err != nil {
		return err
	}

	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, value)

	_, err = McClient.Write(deviceName, offset, 1, data)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", task.TaskName, err)
	}

	// Process the response
	fmt.Printf("Successfully write %s: %d\n", task.TaskName, value)

	// Verify the write operation if needed
	response, err := McClient.Read(deviceName, offset, 1)
	if err != nil {
		return fmt.Errorf("error verifying write for %s: %v", task.TaskName, err)
	}

	if len(response) < 2 {
		return fmt.Errorf("invalid response length during verification")
	}

	readValue := binary.LittleEndian.Uint16(response[11:13])
	//readValue := binary.LittleEndian.Uint16(response)

	if readValue != value {
		return fmt.Errorf("write verification failed for %s: expected %d, got %d", task.TaskName, value, readValue)
	}

	return nil
}
