package conn

import (
	"apcs_refactored/webserver"
	"fmt"
	"net"
	"time"

	mc "github.com/future-architect/go-mcprotocol/mcp"

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
			log.Error(err)
		}
		// TODO - 사업자에게 알림

		//return

	case "물품 끼임":
		log.Infof("[PLC] 물품 끼임")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림

		//return

	case "물품 낙하":
		log.Infof("[PLC] 물품 낙하")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림

		//return

	case "이물질 감지":
		log.Infof("[PLC] 이물질 감지")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림

		//return

	}
}

func InitConnPlc() {
	log.Debugf("plc conn started")

	// go-mcprotocol 라이브러리
	client, err := mc.New3EClient("192.168.50.219", 6000, mc.NewLocalStation())
	if err != nil {
		log.Error("Failed to connect to PLC:", err)
		return
	}

	b := []byte("1")
	// deviceName: device code name 'D' register/ offset: device offset addr/ numPoints: number of read device pointes
	client.Write("D", 100, 3, b)

	go func() {
		for {
			read, err := client.Read("D", 100, 3)
			data := string(read)
			fmt.Println("response:", string(read))
			// registerBinary, _ := mcp.NewParser().Do(read)
			// fmt.Println(string(registerBinary.Payload))

			if err != nil {
				log.Error(err)
				return
			}
			SenseTrouble(data)
			//time.Sleep(10 * time.Millisecond)
			time.Sleep(5 * time.Second)
		}
	}()

	/* // PLC 주소 및 포트 설정
	plcAddress := "192.168.50.219:6000"
	conn, err := net.Dial("tcp", plcAddress)
	if err != nil {
		fmt.Println("Failed to connect to PLC:", err)
		return
	}
	defer conn.Close()

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
	fmt.Println("Received response:", data) */
}
