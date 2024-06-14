package conn

import (
	"apcs_refactored/webserver"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/BRA1L0R/go-mcproto"
	"github.com/BRA1L0R/go-mcproto/packets/models"
	mc "github.com/future-architect/go-mcprotocol/mcp"
	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

// PLC
type PLC struct {
	addr string // 주소
	conn net.Conn
}

type State struct {
	D0 int
	D1 int
}

// MC 프레임 구조체
type MCFrame struct {
	Command byte
	Header  Header
	Packet  Packet
}
type Header struct {
	Protocol byte // 프로토콜 종류
	Address  int  // 주소
	Length   int  // 길이
}

type Packet struct {
	Header   Header // 헤더
	Data     []byte // 데이터
	Checksum byte   // 체크섬
}

// 트러블 감지
func SenseTrouble(data string) {

	/* if state.D0 == 1 {
		SenseTrouble(data)
	} */

	/* go func() {
		fmt.Println("실행")

		waiting <- "화재"
	}() */

	switch data {
	case "화재":
		log.Infof("[PLC] 화재발생")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Println("Error changing view:", err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("화재")
		//return

	case "물품 끼임":
		log.Infof("[PLC] 물품 끼임")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("물품 끼임")
		//return

	case "물품 낙하":
		log.Infof("[PLC] 물품 낙하")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("물품 낙하")

		//return

	case "이물질 감지":
		log.Infof("[PLC] 이물질 감지")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("이물질 감지")

		//return

	}
}

func InitConnPlc0() {
	const addr = "ws://localhost:6000"

	u, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}

	// 웹소켓 연결
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// 메시지 송신
	err = c.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket Server!"))
	if err != nil {
		log.Println("Write error:", err)
		return
	}

	// 메시지 수신 루프
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}
		SenseTrouble(string(msg))

		log.Printf("Received message: %s\n", msg)
	}
}

func InitConnPlc1() {
	// "github.com/BRA1L0R/go-mcproto"
	// "github.com/BRA1L0R/go-mcproto/packets/models"

	client := mcproto.Client{}
	client.Initialize("192.168.50.219", 6000, 755, "apcs")

	for {
		packet, err := client.ReceivePacket()
		if err != nil {
			log.Error(err)
		}

		log.Println("ReceivePacket: ", packet)

		if packet.PacketID == 0x1F { // clientbound keepalive packetid
			receivedKeepalive := new(models.KeepAlivePacket)
			err := packet.DeserializeData(receivedKeepalive)
			if err != nil {
				log.Error(err)
			}

			serverBoundKeepalive := new(models.KeepAlivePacket)
			serverBoundKeepalive.KeepAliveID = receivedKeepalive.KeepAliveID
			serverBoundKeepalive.PacketID = 0x10 // serverbound keepaliveid

			log.Println("ReceivePacket: ", serverBoundKeepalive)

			err = client.WritePacket(serverBoundKeepalive)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func InitConnPlc2() {
	//	mc "github.com/future-architect/go-mcprotocol/mcp"

	log.Debugf("plc conn started")

	// go-mcprotocol 라이브러리
	client, err := mc.New3EClient("192.168.50.219", 6000, mc.NewLocalStation())
	log.Println("connect to PLC:", client)

	if err != nil {
		log.Error("Failed to connect to PLC:", err)
		return
	}

	b := []byte("1")
	// deviceName: device code name 'D' register/ offset: device offset addr/ numPoints: number of read device pointes
	client.Write("D", 1000, 3, b)

	go func() {
		for {
			read, err := client.Read("D", 1000, 3)
			data := string(read)
			//fmt.Println("response:", string(read))
			// registerBinary, _ := mcp.NewParser().Do(read)
			// fmt.Println(string(registerBinary.Payload))

			if err != nil {
				log.Error(err)
				// 알림
				return
			}
			SenseTrouble(data)
			//time.Sleep(100 * time.Millisecond)
			time.Sleep(1 * time.Second)
		}
	}()
}

func InitConnPlc3() {
	// PLC 주소 및 포트 설정
	plcAddress := "localhost:6000"
	conn, err := net.Dial("tcp", plcAddress)
	if err != nil {
		fmt.Println("Failed to connect to PLC:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to PLC")

	// MC 프레임 생성
	frame := MCFrame{
		Command: 68,
		// 다른 필드 초기화
	}

	// MC 프레임을 PLC로 전송
	_, err = conn.Write([]byte{frame.Command}) // 예시: 실제 프레임 전송 방식 사용
	if err != nil {
		fmt.Println("Failed to send MC frame:", err)
		return
	}

	// PLC로부터 응답 수신 및 처리
	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		fmt.Println("Failed to read response:", err)
		return
	}

	// 응답 데이터 처리
	// 실제 MC 프로토콜에 따라 데이터 파싱
	data := string(response)
	fmt.Println("Received response:", data)
}

func InitConnPlc4() {
	conn, err := net.Dial("tcp", "localhost:6000")
	if err != nil {
		log.Printf("Failed to connect to PLC: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to PLC")

	// 쓰기 예제
	writeData := []byte{0x01, 0x00, 0x00}
	log.Println("writeData", writeData)

	_, err = conn.Write(append([]byte{0x14, 0x01, 0x00, 0x00}, writeData...))
	if err != nil {
		log.Printf("Write error: %v", err)
	}

	go func() {
		for {
			// 읽기 예제
			_, err := conn.Write([]byte{0x04, 0x01, 0x00, 0x00, 0x00, 0x03})
			if err != nil {
				log.Printf("Read request error: %v", err)
			}

			readBuffer := make([]byte, 1024)
			n, err := conn.Read(readBuffer)
			if err != nil {
				log.Printf("Read error: %v", err)
			}
			log.Printf("Data read from PLC: %v\n", readBuffer[:n])
			time.Sleep(1 * time.Second)
		}
	}()

	// Keep the main function running
	select {}
}

// func InitConnPlc5() {
// 	log.Println("plc conn started")

// 	client, err := mcproto.New3EClient("localhost", 6000, mcproto.NewLocalStation())
// 	if err != nil {
// 		log.Fatal("Failed to connect to PLC:", err)
// 		return
// 	}
// 	defer client.Close()

// 	log.Println("Connected to PLC:", client)

// 	// 쓰기 예제
// 	writeData := []byte{0x01, 0x00, 0x00}
// 	err = client.Write("D", 1000, 3, writeData)
// 	if err != nil {
// 		log.Fatal("Write error:", err)
// 	}

// 	go func() {
// 		for {
// 			// 읽기 예제
// 			read, err := client.Read("D", 1000, 3)
// 			if err != nil {
// 				log.Fatal("Read error:", err)
// 			}
// 			log.Println("Data read from PLC:", read)
// 			time.Sleep(1 * time.Second)
// 		}
// 	}()

// 	// Keep the main function running
// 	select {}
// }

func CreateMcProtocolMessage() []byte {
	// MC 프로토콜 메시지 작성
	return []byte{
		0x50, 0x00, // Subheader
		0x00, 0xFF, // Network number, PC number
		0xFF, 0x03, 0x00, // Request destination module I/O, station number
		0x0A, 0x00, // Request data length
		0x10, 0x00, // CPU monitoring timer
		0x01, 0x04, // Command (read)
		0x00, 0x00, // Subcommand
		0x00, 0x00, 0x00, // Starting address
		0x00, 0x10, // Number of points
	}
}
