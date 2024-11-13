package conn

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
)

// PLC 메모리 주소 정의 (D5500 ~ D5505)
const (
	// D5500 (전면부 명령) - Modbus 주소 = 실제 주소 - 40001
	FRONT_ADDR        = 5500 // D5500
	FRONT_X_MOVE      = 0    // X축 이동 지령 비트
	FRONT_Z_MOVE      = 1    // Z축 이동 지령 비트
	FRONT_HANDLER_GET = 2    // Z축 핸들러 GET 지령 비트
	FRONT_HANDLER_PUT = 3    // Z축 핸들러 PUT 지령 비트

	// D5501 (후면부 명령) - Modbus 주소 = 실제 주소 - 40001
	REAR_ADDR        = 5501 // D5501
	REAR_X_MOVE      = 0
	REAR_Z_MOVE      = 1
	REAR_HANDLER_GET = 2
	REAR_HANDLER_PUT = 3

	// D5502 (도어 제어) - Modbus 주소 = 실제 주소 - 40001
	DOOR_ADDR         = 5502 // D5502
	IN_DOOR_OPEN      = 0
	IN_DOOR_CLOSE     = 1
	OUT_DOOR_OPEN     = 2
	OUT_DOOR_CLOSE    = 3
	ROBOT_DOOR1_OPEN  = 4
	ROBOT_DOOR1_CLOSE = 5
	ROBOT_DOOR2_OPEN  = 6
	ROBOT_DOOR2_CLOSE = 7

	// D5503 (운전 제어) - Modbus 주소 = 실제 주소 - 40001
	OPERATION_ADDR   = 5503 // D5503
	OP_HOME_POSITION = 0    // 대기위치 이동
	OP_START         = 1    // 운전 시작
	OP_STOP          = 2    // 운전 정지
	OP_EMERGENCY     = 3    // 비상정지

	// D5505 (알람) - Modbus 주소 = 실제 주소 - 40001
	ALARM_ADDR = 5505 // D5505
)

var (
	modbusClient modbus.Client
	handler      *modbus.TCPClientHandler
)

func InitModbusClient() error {
	// TCP 핸들러 생성
	handler = modbus.NewTCPClientHandler("localhost:502")
	// handler = modbus.NewTCPClientHandler("192.168.0.10:5000")  // PLC IP 주소

	// 고급 설정
	handler.Timeout = 10 * time.Second
	handler.SlaveId = 0xFF // 1

	// 연결
	if err := handler.Connect(); err != nil {
		return err
	}

	modbusClient = modbus.NewClient(handler)
	log.Info("[PLC] Modbus 연결 완료")
	return nil
}

// 연결 종료
func CloseModbusClient() {
	if handler != nil {
		handler.Close()
	}
}

// 레지스터 읽기
func ReadRegister(address uint16, quantity uint16) ([]byte, error) {
	// D5500을 Modbus 주소로 변환해서 읽기
	results, err := modbusClient.ReadHoldingRegisters(address, quantity)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// 레지스터 쓰기 (단일)
func WriteRegister(address uint16, value uint16) error {
	_, err := modbusClient.WriteSingleRegister(address, value)
	if err != nil {
		log.Errorf("[PLC] Failed to write register: %v", err)
		return err
	}
	return nil
}

// 레지스터 쓰기 (다중)
func WriteMultipleRegisters(address uint16, quantity uint16, values []byte) error {
	_, err := modbusClient.WriteMultipleRegisters(address, quantity, values)
	if err != nil {
		log.Errorf("[PLC] Failed to write multiple registers: %v", err)
		return err
	}
	return nil
}

// 명령 실행 및 완료 대기
func ExecuteCommand(address uint16, bitPosition uint) error {
	// 1. 현재 값이 이미 1인지 확인
	result, err := modbusClient.ReadHoldingRegisters(address, 1)
	if err != nil {
		return err
	}
	currentValue := binary.BigEndian.Uint16(result)

	// 이미 1이면 에러
	if (currentValue & (1 << bitPosition)) != 0 {
		return fmt.Errorf("bit already set: addr=%d, bit=%d", address, bitPosition)
	}

	// 2. 비트 설정
	newValue := currentValue | (1 << bitPosition)
	_, err = modbusClient.WriteSingleRegister(address, newValue)
	if err != nil {
		return err
	}

	// 3. 완료 대기 (최대 3초로 수정)
	timeout := time.After(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("command timeout: addr=%d, bit=%d", address, bitPosition)
		case <-ticker.C:
			result, err := modbusClient.ReadHoldingRegisters(address, 1)
			if err != nil {
				return err
			}
			currentValue = binary.BigEndian.Uint16(result)

			if (currentValue & (1 << bitPosition)) == 0 {
				log.Info("[PLC_Command] 명령 완료")
				return nil
			}
		}
	}
}

// 전면부 X축 이동
func MoveFrontX() error {
	return ExecuteCommand(FRONT_ADDR, FRONT_X_MOVE)
}

// 전면부 핸들러 GET
func FrontHandlerGet() error {
	return ExecuteCommand(FRONT_ADDR, FRONT_HANDLER_GET)
}

// 입구 도어 열기
func OpenInDoor() error {
	return ExecuteCommand(DOOR_ADDR, IN_DOOR_OPEN)
}

// 알람 상태 확인
func CheckAlarms() (uint16, error) {
	result, err := modbusClient.ReadHoldingRegisters(ALARM_ADDR, 1)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(result), nil
}

// 비트 해석은 PLC의 메모리 맵에 따라
func CheckBit(value uint16, bitPosition uint) bool {
	return (value & (1 << bitPosition)) != 0
}
