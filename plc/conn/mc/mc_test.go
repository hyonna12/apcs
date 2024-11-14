package mc

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMCProtocol(t *testing.T) {
	log.SetLevel(log.DebugLevel) // 로그 레벨 설정

	// 1. 클라이언트 생성
	// client, err := NewMCClient("192.168.0.10", 5000, MODE_BINARY)
	client, err := NewMCClient("192.168.0.10", 5000, MODE_ASCII)
	// client, err := NewMCClient("localhost", 5000)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 2. 연결
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 3. D5000 영역 읽기 (10개)
	log.Info("1. D5000 영역 읽기 (10개)")
	values, err := client.ReadDRegister(5000, 10)
	if err != nil {
		t.Errorf("Failed to read D5000: %v", err)
		return
	}

	// 읽은 값 출력 (다양한 형식)
	for i, value := range values {
		// 모든 형식으로 출력
		log.Infof("D%d = %s", 5000+i, value.ToString())

		// ON 상태인 비트 출력
		for bit := uint(0); bit < 16; bit++ {
			if bitVal, err := value.GetBit(bit); err == nil && bitVal {
				log.Infof("  Bit %d: ON", bit)
			}
		}
	}

	// 4. D5500에 값 쓰기
	log.Info("\n2. D5500에 값 쓰기")
	err = client.WriteDRegister(5500, 0x1234)
	if err != nil {
		t.Errorf("Failed to write D5500: %v", err)
		return
	}

	// 5. 쓴 값 확인
	log.Info("\n3. D5500 값 확인")
	values, err = client.ReadDRegister(5500, 1)
	if err != nil {
		t.Errorf("Failed to read D5500: %v", err)
		return
	}

	if len(values) > 0 {
		val, err := values[0].ToUint16()
		if err != nil {
			t.Errorf("Failed to convert value: %v", err)
			return
		}
		log.Infof("D5500 = %016b (0x%04X, DEC: %s)",
			val,
			val,
			values[0].ToDecimal())
	}
}

func Test(t *testing.T) {
	// 1. 클라이언트 생성
	client, err := NewMCClient("192.168.0.10", 5000, MODE_ASCII)
	// client, err := NewMCClient("localhost", 5000)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 2. 연결
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	log.Info("1. 읽기")
	err = client.CheckConnection()
	if err != nil {
		t.Errorf("Failed to read D5000: %v", err)
		return
	}

}

// 2워드 쓰기 테스트 수정
func TestWriteTwoWords(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	client, err := NewMCClient("192.168.0.10", 5000, MODE_ASCII)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 2워드 쓰기 테스트
	log.Info("1. D5500-D5501에 2워드 쓰기")
	err = client.WriteDRegister2Word(5500, 0x1234, 0x5678)
	if err != nil {
		t.Errorf("Failed to write 2 words: %v", err)
		return
	}

	// 쓴 값 확인
	log.Info("2. D5500-D5501 값 확인")
	values, err := client.ReadDRegister(5000, 2)
	if err != nil {
		t.Errorf("Failed to read values: %v", err)
		return
	}

	if len(values) >= 2 {
		val1, err := values[0].ToUint16()
		if err != nil {
			t.Errorf("Failed to convert first value: %v", err)
			return
		}
		val2, err := values[1].ToUint16()
		if err != nil {
			t.Errorf("Failed to convert second value: %v", err)
			return
		}

		log.Infof("D5500 = %s (Value: 0x%04X)", values[0].ToString(), val1)
		log.Infof("D5501 = %s (Value: 0x%04X)", values[1].ToString(), val2)

		// 예상 값과 비교
		if val1 != 0x1234 || val2 != 0x5678 {
			t.Errorf("Unexpected values: D5500=0x%04X (expected 0x1234), D5501=0x%04X (expected 0x5678)",
				val1, val2)
		}
	}
}

func TestReadPLCMemory(t *testing.T) {
	client, err := NewMCClient("192.168.0.10", 5000, MODE_ASCII)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 1. 2워드 값 읽기 (D5500-D5501)
	log.Info("1. D5500-D5501 32비트 값 읽기")
	value, err := client.ReadDRegister2Word(5500)
	if err != nil {
		t.Errorf("Failed to read 32-bit value: %v", err)
		return
	}
	log.Infof("D5500-D5501 = %s", value.ToString32())

	// 2. 1워드 값 읽기 (D5510)
	log.Info("2. D5510 16비트 값 읽기")
	values, err := client.ReadDRegister(5510, 1)
	if err != nil {
		t.Errorf("Failed to read 16-bit value: %v", err)
		return
	}
	if len(values) > 0 {
		log.Infof("D5510 = %s", values[0].ToString())
	}
}

func TestWritePLCMemory(t *testing.T) {
	client, err := NewMCClient("192.168.0.10", 5000, MODE_ASCII)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 1. 32비트 값(1234567) 쓰기
	value := uint32(12345678) // 0x0012D687
	log.Infof("1. D5500-D5501에 32비트 값 쓰기: %d (0x%08X)", value, value)

	err = client.WriteDRegister32(5500, value)
	if err != nil {
		t.Errorf("Failed to write 32-bit value: %v", err)
		return
	}

	// 2. 쓴 값 확인
	log.Info("2. D5500-D5501 값 확인")
	readValue, err := client.ReadDRegister2Word(5500)
	if err != nil {
		t.Errorf("Failed to read back 32-bit value: %v", err)
		return
	}

	// 읽은 값 출력 (다양한 형식)
	uint32Val, _ := readValue.ToUint32()
	log.Infof("읽은 값: %s", readValue.ToString32())

	// 예상 값과 비교
	if uint32Val != value {
		t.Errorf("Unexpected value: got 0x%08X, want 0x%08X", uint32Val, value)
	}

	// 3. 개별 워드 값 확인
	values, err := client.ReadDRegister(5500, 2)
	if err != nil {
		t.Errorf("Failed to read individual words: %v", err)
		return
	}

	if len(values) >= 2 {
		// 상위 워드 (D5500)
		highWord, _ := values[0].ToUint16()
		// 하위 워드 (D5501)
		lowWord, _ := values[1].ToUint16()

		log.Infof("D5500 (상위 워드) = %s", values[0].ToString())
		log.Infof("D5501 (하위 워드) = %s", values[1].ToString())

		// 예상되는 상위/하위 워드 값과 비교
		expectedHigh := uint16(value >> 16)   // 0x0012
		expectedLow := uint16(value & 0xFFFF) // 0xD687

		if highWord != expectedHigh || lowWord != expectedLow {
			t.Errorf("Unexpected word values: D5500=0x%04X (expected 0x%04X), D5501=0x%04X (expected 0x%04X)",
				highWord, expectedHigh, lowWord, expectedLow)
		}
	}
}

func TestModifyFirstTwoWords(t *testing.T) {
	client, err := NewMCClient("192.168.0.10", 5000, MODE_ASCII)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 1. 현재 값 읽기
	log.Info("1. 현재 D5500-D5501 값 읽기")
	currentValue, err := client.ReadDRegister2Word(5500)
	if err != nil {
		t.Errorf("Failed to read current value: %v", err)
		return
	}
	log.Infof("현재 값: %s", currentValue.ToString32())

	// 2. 새로운 값(12345678) 쓰기
	newValue := uint32(12345678) // 0x00BC614E
	log.Infof("2. D5500-D5501에 새 값 쓰기: %d (0x%08X)", newValue, newValue)

	err = client.WriteDRegister32(5500, newValue)
	if err != nil {
		t.Errorf("Failed to write new value: %v", err)
		return
	}

	// 3. 변경된 값 확인
	log.Info("3. 변경된 값 확인")
	updatedValue, err := client.ReadDRegister2Word(5500)
	if err != nil {
		t.Errorf("Failed to read updated value: %v", err)
		return
	}

	// 변경된 값 출력
	uint32Val, _ := updatedValue.ToUint32()
	log.Infof("변경된 값: %s", updatedValue.ToString32())

	// 예상 값과 비교
	if uint32Val != newValue {
		t.Errorf("Unexpected value: got 0x%08X, want 0x%08X", uint32Val, newValue)
	}

	// 4. 개별 워드 값 확인
	values, err := client.ReadDRegister(5500, 2)
	if err != nil {
		t.Errorf("Failed to read individual words: %v", err)
		return
	}

	if len(values) >= 2 {
		highWord, _ := values[0].ToUint16() // D5500 (상위 워드)
		lowWord, _ := values[1].ToUint16()  // D5501 (하위 워드)

		log.Infof("D5500 (상위 워드) = %s", values[0].ToString())
		log.Infof("D5501 (하위 워드) = %s", values[1].ToString())

		// 예상되는 상위/하위 워드 값
		expectedHigh := uint16(newValue >> 16)   // 0x00BC
		expectedLow := uint16(newValue & 0xFFFF) // 0x614E

		if highWord != expectedHigh || lowWord != expectedLow {
			t.Errorf("Unexpected word values: D5500=0x%04X (expected 0x%04X), D5501=0x%04X (expected 0x%04X)",
				highWord, expectedHigh, lowWord, expectedLow)
		}
	}
}
