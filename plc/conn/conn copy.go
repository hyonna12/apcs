package conn

// import (
// 	"apcs_refactored/webserver"
// 	"encoding/binary"
// 	"net"
// 	"time"

// 	mc "github.com/future-architect/go-mcprotocol/mcp"

// 	log "github.com/sirupsen/logrus"
// )

// // PLC
// type PLC struct {
// 	addr string // 주소
// 	conn net.Conn
// }

// type State struct {
// 	D0 int
// 	D1 int
// 	D2 int
// }

// // MC 프레임 구조체
// type MCFrame struct {
// 	SubHeader       [4]byte
// 	AccessRoute     [2]byte
// 	DataLength      [2]byte
// 	MonitoringTimer [2]byte
// 	Command         [2]byte
// 	SubCommand      [2]byte
// 	Data            []byte
// }

// type Header struct {
// 	Protocol byte // 프로토콜 종류
// 	Address  int  // 주소
// 	Length   int  // 길이
// }

// type Packet struct {
// 	Header   Header // 헤더
// 	Data     []byte // 데이터
// 	Checksum byte   // 체크섬
// }

// // 트러블 감지
// func SenseTrouble(data string) {
// 	log.Println(data)

// 	/* if state.D0 == 1 {
// 		SenseTrouble(data)
// 	} */

// 	/* go func() {
// 		fmt.Println("실행")

// 		waiting <- "화재"
// 	}() */

// 	switch data {
// 	case "화재":
// 		log.Infof("[PLC] 화재발생")
// 		err := webserver.ChangeKioskView("/error/trouble")
// 		if err != nil {
// 			log.Println("Error changing view:", err)
// 		}
// 		// TODO - 사업자에게 알림
// 		webserver.SendEvent("화재")
// 		//return

// 	case "물품 끼임":
// 		log.Infof("[PLC] 물품 끼임")
// 		err := webserver.ChangeKioskView("/error/trouble")
// 		if err != nil {
// 			log.Error(err)
// 		}
// 		// TODO - 사업자에게 알림
// 		webserver.SendEvent("물품 끼임")
// 		//return

// 	case "물품 낙하":
// 		log.Infof("[PLC] 물품 낙하")
// 		err := webserver.ChangeKioskView("/error/trouble")
// 		if err != nil {
// 			log.Error(err)
// 		}
// 		// TODO - 사업자에게 알림
// 		webserver.SendEvent("물품 낙하")

// 		//return

// 	case "이물질 감지":
// 		log.Infof("[PLC] 이물질 감지")
// 		err := webserver.ChangeKioskView("/error/trouble")
// 		if err != nil {
// 			log.Error(err)
// 		}
// 		// TODO - 사업자에게 알림
// 		webserver.SendEvent("이물질 감지")

// 		//return

// 	}
// }

// func InitConnPlc() {
// 	//	mc "github.com/future-architect/go-mcprotocol/mcp"

// 	log.Debugf("plc conn started")

// 	client, err := mc.New3EClient("localhost", 6060, mc.NewLocalStation())

// 	if err != nil {
// 		log.Fatal("Failed to connect to PLC:", err)
// 		return
// 	}

// 	log.Println("connect to PLC:", client)

// 	// 초기 상태 쓰기
// 	err = WriteToPLC(client, 1000, []uint16{0, 0, 0})
// 	if err != nil {
// 		log.Error("Failed to write initial state to PLC:", err)
// 		return
// 	}

// 	// 주기적인 읽기 작업을 위한 타이머 설정
// 	ticker := time.NewTicker(1 * time.Second)
// 	go func() {
// 		for range ticker.C {
// 			ReadFromPLC(client)
// 		}
// 	}()

// 	// go func() {
// 	// 	for {
// 	// 		read, err := client.Read("D", 1000, 3)
// 	// 		data := string(read)
// 	// 		//fmt.Println("response:", string(read))
// 	// 		// registerBinary, _ := mcp.NewParser().Do(read)
// 	// 		// fmt.Println(string(registerBinary.Payload))

// 	// 		if err != nil {
// 	// 			log.Error(err)
// 	// 			// 알림
// 	// 			return
// 	// 		}
// 	// 		SenseTrouble(data)
// 	// 		//time.Sleep(100 * time.Millisecond)
// 	// 		time.Sleep(1 * time.Second)
// 	// 	}
// 	// }()
// }

// func WriteToPLC(client mc.Client, startAddress uint32, values []uint16) error {
// 	data := make([]byte, len(values)*2)
// 	for i, v := range values {
// 		binary.LittleEndian.PutUint16(data[i*2:], v)
// 	}

// 	_, err := client.Write("D", int64(startAddress), int64(len(values)), data)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func ReadFromPLC(client mc.Client) {
// 	read, err := client.Read("D", int64(1000), int64(3))
// 	if err != nil {
// 		log.Fatal("Failed to read from PLC:", err)
// 		return
// 	}

// 	if len(read) >= 6 {
// 		state := State{
// 			D0: int(binary.LittleEndian.Uint16(read[0:2])),
// 			D1: int(binary.LittleEndian.Uint16(read[2:4])),
// 			D2: int(binary.LittleEndian.Uint16(read[4:6])),
// 		}

// 		log.Printf("Read state: D0=%d, D1=%d, D2=%d", state.D0, state.D1, state.D2)

// 		// 상태에 따른 처리
// 		if state.D0 != 0 || state.D1 != 0 || state.D2 != 0 {
// 			troubleState := string([]rune{rune(state.D0), rune(state.D1), rune(state.D2)})
// 			SenseTrouble(troubleState)
// 		}
// 	} else {
// 		log.Error("Insufficient data read from PLC")
// 	}
// }
