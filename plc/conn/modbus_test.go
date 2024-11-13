package conn

import (
	"encoding/binary"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestModbusCommunication(t *testing.T) {
	err := InitModbusClient()
	if err != nil {
		t.Fatalf("Failed to initialize Modbus client: %v", err)
	}
	defer CloseModbusClient()

	// 1. D5000 영역 읽기 (400개를 125개씩 나눠서)
	log.Info("1. D5000 영역 읽기 (400개)")
	var allResults []byte

	// 125개씩 4번 읽기
	for i := 0; i < 400; i += 125 {
		count := uint16(125)
		if i+125 > 400 {
			count = uint16(400 - i)
		}

		result, err := ReadRegister(uint16(5000+i), count)
		if err != nil {
			t.Errorf("Failed to read D%d: %v", 5000+i, err)
			return
		}
		allResults = append(allResults, result...)
	}

	// 읽은 값 중 처음 10개만 출력
	for i := 0; i < 20 && i < len(allResults); i += 2 {
		value := binary.BigEndian.Uint16(allResults[i : i+2])
		log.Infof("D%d = %016b", 5000+i/2, value)
	}

	// 2. D5500 영역 쓰기 (10개)
	log.Info("\n2. D5500 영역 쓰기 (10개)")
	for i := 0; i < 10; i++ {
		addr := uint16(5500 + i)
		value := uint16(1 << (i % 16)) // 각 레지스터마다 다른 비트 설정
		err = WriteRegister(addr, value)
		if err != nil {
			t.Errorf("Failed to write D%d: %v", addr, err)
			return
		}
		log.Infof("D%d <- %016b", addr, value)
	}

	// 3. D5500 영역 읽기 (쓴 값 확인)
	log.Info("\n3. D5500 영역 읽기 (확인)")
	result, err := ReadRegister(5500, 10)
	if err != nil {
		t.Errorf("Failed to read D5500: %v", err)
		return
	}

	for i := 0; i < len(result); i += 2 {
		value := binary.BigEndian.Uint16(result[i : i+2])
		log.Infof("D%d = %016b", 5500+i/2, value)
	}
}

func TestHardwareOperation(t *testing.T) {
	err := InitModbusClient()
	if err != nil {
		t.Fatalf("Failed to initialize Modbus client: %v", err)
	}
	defer CloseModbusClient()

	// 1. 전면부 X축 이동
	err = MoveFrontX()
	if err != nil {
		t.Errorf("Failed to move front X: %v", err)
	}

	// 2. 입구 도어 열기
	err = OpenInDoor()
	if err != nil {
		t.Errorf("Failed to open in door: %v", err)
	}

	// 3. 알람 상태 확인
	alarms, err := CheckAlarms()
	if err != nil {
		t.Errorf("Failed to check alarms: %v", err)
	}
	log.Infof("Alarm status: %016b", alarms)
}

func TestInboundProcess(t *testing.T) {
	err := InitModbusClient()
	if err != nil {
		t.Fatalf("Failed to initialize Modbus client: %v", err)
	}
	defer CloseModbusClient()

	timeout := time.After(30 * time.Second)
	done := make(chan bool)

	go func() {
		// 1. 입구 도어 열기 (D5502.0 = 1 설정, D5002.0 완료 확인)
		log.Info("1. 입구 도어 열기 시작")
		err = WriteRegister(5502, 0x0001) // IN_DOOR_OPEN 비트
		if err != nil {
			t.Errorf("도어 열기 명령 실패: %v", err)
			done <- false
			return
		}

		// 완료 대기
		for {
			result, err := ReadRegister(5002, 1)
			if err != nil {
				t.Errorf("도어 상태 읽기 실패: %v", err)
				done <- false
				return
			}
			value := binary.BigEndian.Uint16(result)
			log.Infof("도어 상태 (D5002): %016b", value)
			if CheckBit(value, 0) {
				log.Info("도어 열기 완료")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		// 2. 전면부 X축 이동 (D5500.0 = 1 설정, D5000.0 완료 확인)
		log.Info("2. 전면부 X축 이동 시작")
		err = WriteRegister(5500, 0x0001) // FRONT_X_MOVE 비트
		if err != nil {
			t.Errorf("X축 이동 명령 실패: %v", err)
			done <- false
			return
		}

		// 완료 대기
		for {
			result, err := ReadRegister(5000, 1)
			if err != nil {
				t.Errorf("X축 상태 읽기 실패: %v", err)
				done <- false
				return
			}
			value := binary.BigEndian.Uint16(result)
			log.Infof("X축 상태 (D5000): %016b", value)
			if CheckBit(value, 0) {
				log.Info("X축 이동 완료")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		// 3. 핸들러로 물품 집기 (D5500.2 = 1 설정, D5000.2 완료 확인)
		log.Info("3. 핸들러로 물품 집기 시작")
		err = WriteRegister(5500, 0x0004) // FRONT_HANDLER_GET 비트
		if err != nil {
			t.Errorf("핸들러 GET 명령 실패: %v", err)
			done <- false
			return
		}

		// 완료 대기
		for {
			result, err := ReadRegister(5000, 1)
			if err != nil {
				t.Errorf("핸들러 상태 읽기 실패: %v", err)
				done <- false
				return
			}
			value := binary.BigEndian.Uint16(result)
			log.Infof("핸들러 상태 (D5000): %016b", value)
			if CheckBit(value, 2) {
				log.Info("물품 집기 완료")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		log.Info("입고 프로세스 완료")
		done <- true
	}()

	// 타임아웃 처리
	select {
	case <-timeout:
		t.Error("입고 프로세스 타임아웃")
	case success := <-done:
		if !success {
			t.Error("입고 프로세스 실패")
		} else {
			log.Info("입고 프로세스 테스트 성공")
		}
	}
}

// go test -v -count=1 -run TestInboundProcess
