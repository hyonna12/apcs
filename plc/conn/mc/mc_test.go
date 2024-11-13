package mc

import (
	"encoding/binary"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMCProtocol(t *testing.T) {
	// 1. 클라이언트 생성
	// client, err := NewMCClient("192.168.0.10", 5000)
	client, err := NewMCClient("localhost", 5000)

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
	result, err := client.ReadDRegister(5000, 10)
	if err != nil {
		t.Errorf("Failed to read D5000: %v", err)
		return
	}

	// 읽은 값 출력
	for i := 0; i < len(result); i += 2 {
		value := binary.BigEndian.Uint16(result[i : i+2])
		log.Infof("D%d = %016b", 5000+i/2, value)
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
	result, err = client.ReadDRegister(5500, 1)
	if err != nil {
		t.Errorf("Failed to read D5500: %v", err)
		return
	}
	value := binary.BigEndian.Uint16(result)
	log.Infof("D5500 = %016b (0x%04X)", value, value)
}

// go test -v ./plc/conn/mc -run TestMCProtocol
