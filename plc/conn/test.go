package conn

// import (
// 	"encoding/binary"
// 	"fmt"
// 	"net"

// 	log "github.com/sirupsen/logrus"
// )

// type PCClient struct {
// 	conn net.Conn
// }

// type PLCData struct {
// 	TaskName string
// 	Address  uint32
// 	Value    uint16
// 	IsRead   bool
// }

// var taskList = map[string]PLCData{
// 	"OpenDoor":       {"OpenDoor", 100, 1, false},
// 	"CloseDoor":      {"CloseDoor", 101, 0, false},
// 	"TurnOnLight":    {"TurnOnLight", 102, 1, false},
// 	"TurnOffLight":   {"TurnOffLight", 103, 0, false},
// 	"SetTemperature": {"SetTemperature", 104, 25, false},
// 	"DetectFire":     {"DetectFire", 105, 0, true},
// 	"Weight":         {"Weight", 106, 0, true},
// 	"Height":         {"Height", 107, 0, true},
// }

// func NewPCClient(address string) (*PCClient, error) {
// 	conn, err := net.Dial("tcp", address)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &PCClient{conn: conn}, nil
// }

// func (pc *PCClient) Close() {
// 	pc.conn.Close()
// }

// func (pc *PCClient) WriteDataRegister(address uint16, value uint16) error {
// 	command := make([]byte, 21)
// 	copy(command, []byte{
// 		0x50, 0x00, // 서브헤더
// 		0x00,       // 네트워크 번호
// 		0xFF,       // PC 번호
// 		0xFF, 0x03, // 요청 대상 모듈 I/O 번호
// 		0x00,       // 요청 대상 모듈 국번
// 		0x0E, 0x00, // 요청 데이터 길이 (14바이트)
// 		0x14, 0x01, // 쓰기 명령
// 		0x00, // 서브 명령
// 		'D',  // 디바이스 코드
// 	})

// 	// 주소 설정 (Big Endian)
// 	binary.BigEndian.PutUint16(command[13:15], address)

// 	// 디바이스 점수 설정 (1개)
// 	binary.BigEndian.PutUint16(command[15:17], 1)

// 	// 데이터 추가 (Little Endian)
// 	binary.LittleEndian.PutUint16(command[17:19], value)

// 	log.Printf("Sending Write command: Address %d, Value %d\n", address, value)
// 	log.Printf("Command bytes: %X\n", command)

// 	_, err := pc.conn.Write(command)
// 	if err != nil {
// 		return err
// 	}

// 	// 응답 읽기
// 	response := make([]byte, 11)
// 	_, err = pc.conn.Read(response)
// 	if err != nil {
// 		return err
// 	}

// 	log.Printf("Received response: %X\n", response)

// 	if response[0] != 0xD0 || response[10] != 0x00 {
// 		return fmt.Errorf("error response: %02X", response[10])
// 	}

// 	return nil
// }

// func (pc *PCClient) ReadDataRegister(address uint16, count uint16) ([]uint16, error) {
// 	command := make([]byte, 17)
// 	copy(command, []byte{
// 		0x50, 0x00, // 서브헤더
// 		0x00,       // 네트워크 번호
// 		0xFF,       // PC 번호
// 		0xFF, 0x03, // 요청 대상 모듈 I/O 번호
// 		0x00,       // 요청 대상 모듈 국번
// 		0x0B, 0x00, // 요청 데이터 길이 (11바이트)
// 		0x04, 0x01, // 읽기 명령
// 		0x00, // 서브 명령
// 		'D',  // 디바이스 코드
// 	})

// 	// 주소 설정 (Big Endian)
// 	binary.BigEndian.PutUint16(command[13:15], address)

// 	// 디바이스 점수 설정
// 	binary.BigEndian.PutUint16(command[15:17], count)

// 	log.Printf("Sending Read command: Address %d, Count %d\n", address, count)
// 	log.Printf("Command bytes: %X\n", command)

// 	_, err := pc.conn.Write(command)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 응답 읽기
// 	header := make([]byte, 11)
// 	_, err = pc.conn.Read(header)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if header[0] != 0xD0 || header[10] != 0x00 {
// 		return nil, fmt.Errorf("error response: %02X", header[10])
// 	}

// 	dataLength := binary.BigEndian.Uint16(header[7:9])
// 	data := make([]byte, dataLength)
// 	_, err = pc.conn.Read(data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Printf("Received response header: %X\n", header)
// 	log.Printf("Received response data: %X\n", data)

// 	values := make([]uint16, count)
// 	for i := uint16(0); i < count; i++ {
// 		values[i] = binary.LittleEndian.Uint16(data[i*2 : (i+1)*2])
// 	}

// 	return values, nil
// }

// func (pc *PCClient) ExecuteTask(taskName string) error {
// 	log.Println(taskName, "실행")
// 	task, exists := taskList[taskName]
// 	if !exists {
// 		return fmt.Errorf("task %s not found", taskName)
// 	}

// 	if task.IsRead {
// 		// 읽기 작업 수행
// 		values, err := pc.ReadDataRegister(uint16(task.Address), 1)
// 		if err != nil {
// 			return fmt.Errorf("error reading %s (Address: %d): %v", taskName, task.Address, err)
// 		}
// 		log.Printf("Read value for %s (Address: %d): %v", taskName, task.Address, values[0])
// 	} else {
// 		// 쓰기 작업 수행
// 		err := pc.WriteDataRegister(uint16(task.Address), task.Value)
// 		if err != nil {
// 			return fmt.Errorf("error writing %s (Address: %d, Value: %d): %v", taskName, task.Address, task.Value, err)
// 		}
// 		log.Printf("Successfully wrote %s (Address: %d, Value: %d)", taskName, task.Address, task.Value)
// 	}

// 	return nil
// }

// func InitConn() {
// 	client, err := NewPCClient("localhost:6060")
// 	if err != nil {
// 		log.Printf("Error connecting to PLC: %v\n", err)
// 		return
// 	}
// 	defer client.Close()

// 	client.ExecuteTask("OpenDoor")
// 	client.ExecuteTask("Height")
// }
