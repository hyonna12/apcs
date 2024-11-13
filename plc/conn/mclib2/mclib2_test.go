package mclib2

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMCLib2Protocol(t *testing.T) {
	// 1. 클라이언트 생성
	client, err := NewMCLib2Client("localhost", 5000)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 2. D5000 영역 읽기 (10개)
	log.Info("1. D5000 영역 읽기 (10개)")
	result, err := client.ReadDRegister(5000, 10)
	if err != nil {
		t.Errorf("Failed to read D5000: %v", err)
		return
	}

	// 읽은 값 출력
	for i, value := range result {
		log.Infof("D%d = %016b", 5000+i, value)
	}

	// 3. D5500에 값 쓰기
	log.Info("\n2. D5500에 값 쓰기")
	err = client.WriteDRegister(5500, 0x1234)
	if err != nil {
		t.Errorf("Failed to write D5500: %v", err)
		return
	}

	// 4. 쓴 값 확인
	log.Info("\n3. D5500 값 확인")
	result, err = client.ReadDRegister(5500, 1)
	if err != nil {
		t.Errorf("Failed to read D5500: %v", err)
		return
	}
	log.Infof("D5500 = %016b (0x%04X)", result[0], result[0])
}

//go test -v ./plc/conn/mclib2 -run TestMCLib2Protocol
