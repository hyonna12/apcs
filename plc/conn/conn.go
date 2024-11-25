package conn

import (
	"apcs_refactored/event/trouble"
	"encoding/json"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type PLCCommand struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

type PLCResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type PLCClient struct {
	conn           net.Conn
	host           string
	port           string
	connected      bool
	troubleHandler trouble.TroubleEventHandler
}

var client *PLCClient

func NewPLCClient(host string, port string) *PLCClient {
	return &PLCClient{
		host: host,
		port: port,
	}
}

func (c *PLCClient) Connect() error {
	conn, err := net.Dial("tcp", c.host+":"+c.port)
	if err != nil {
		return err
	}
	c.conn = conn
	c.connected = true

	// Keep-alive 설정
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(30 * time.Second)

	return nil
}

func (c *PLCClient) SendCommand(cmd *PLCCommand) (*PLCResponse, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to PLC server")
	}

	// 타임아웃 설정 증가
	c.conn.SetDeadline(time.Now().Add(30 * time.Second))

	// 재시도 로직
	for retries := 0; retries < 3; retries++ {
		data, err := json.Marshal(cmd)
		if err != nil {
			return nil, err
		}

		_, err = c.conn.Write(data)
		if err != nil {
			log.Errorf("Send error (retry %d): %v", retries+1, err)
			time.Sleep(time.Second)
			continue
		}

		// 응답 읽기
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			log.Errorf("Read error (retry %d): %v", retries+1, err)
			time.Sleep(time.Second)
			continue
		}

		var response PLCResponse
		err = json.Unmarshal(buf[:n], &response)
		if err != nil {
			return nil, err
		}

		// 응답의 success 필드로 완료 여부 확인
		return &response, nil
	}
	return nil, fmt.Errorf("failed after 3 retries")
}

func ConnectPlcServer() *PLCClient {
	log.Info("[PLC] PLC 서버 연결 시도...")

	client = NewPLCClient("localhost", "5000")

	// 연결 재시도 루프
	for {
		err := client.Connect()
		if err != nil {
			log.Error("Failed to connect to PLC server:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Info("Connected to PLC server")

		// 연결 성공 시 메시지 리스너 시작
		go listenForMessages()
		go client.StartPingPong() // Ping-Pong 시작

		break
	}
	return client
}

// 메시지 리스닝 고루틴
func listenForMessages() {
	log.Info("[PLC] 메시지 리스닝 시작")

	for {
		if client == nil || !client.connected {
			log.Error("[PLC] 서버 연결이 끊어짐")
			time.Sleep(5 * time.Second)
			continue
		}

		// 메시지 수신
		response, err := GetResponse()
		if err != nil {
			log.Errorf("[PLC] 메시지 수신 에러: %v", err)
			continue
		}

		// command_id가 "0"인 경우 트러블 메시지로 처리
		if response.CommandId == "0" {
			handleTroubleMessage(response)
			continue
		}

		// 일반 메시지 로깅
		log.Infof("[PLC] 일반 메시지 수신: %+v", response)
	}
}

func handleTroubleMessage(response *Response) {
	if client.troubleHandler == nil {
		log.Error("[PLC] 트러블 핸들러가 설정되지 않음")
		return
	}

	details := response.Details
	troubleType, ok := details["type"].(string)
	if !ok {
		log.Error("[PLC] 트러블 타입 파싱 실패")
		return
	}

	log.Infof("[PLC] 트러블 메시지 수신: %s", troubleType)
	client.troubleHandler.HandleTroubleEvent(troubleType, details)
}

// Response PLC 서버로부터의 응답 구조체
type Response struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	CommandId string                 `json:"command_id"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// GetResponse PLC 서버로부터 응답을 받는 함수
func GetResponse() (*Response, error) {
	// 응답 데이터 수신
	data, err := readData()
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	// JSON 디코딩
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	return &resp, nil
}

// readData PLC 서버로부터 데이터를 읽는 함수
func readData() ([]byte, error) {
	if client == nil || !client.connected {
		return nil, fmt.Errorf("PLC 서버에 연결되어 있지 않음")
	}

	// 타임아웃 설정 증가
	client.conn.SetDeadline(time.Now().Add(30 * time.Second))

	// 응답 읽기
	buf := make([]byte, 1024)
	n, err := client.conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("데이터 읽기 실패: %v", err)
	}

	return buf[:n], nil
}

// 핸들러 설정 함수 추가
func (c *PLCClient) SetTroubleHandler(handler trouble.TroubleEventHandler) {
	c.troubleHandler = handler
}

// StartPingPong 주기적으로 서버에 ping을 보내는 함수
func (c *PLCClient) StartPingPong() {
	ticker := time.NewTicker(20 * time.Second) // 30초마다 Ping
	defer ticker.Stop()

	for range ticker.C {
		if !c.connected {
			log.Warn("[PLC] Not connected to server, skipping ping")
			continue
		}

		pingCmd := &PLCCommand{Type: "ping"}
		response, err := c.SendCommand(pingCmd)
		if err != nil {
			log.Errorf("[PLC] Ping error: %v", err)
			c.connected = false
			continue
		}

		if response.Message != "pong" {
			log.Warn("[PLC] Unexpected pong response")
		} else {
			log.Info("[PLC] Pong received")
		}
	}
}

// GetPLCClient returns the current PLC client instance
func GetPLCClient() *PLCClient {
	if client == nil {
		log.Error("[PLC] PLC client is not initialized")
		return nil
	}
	return client
}
