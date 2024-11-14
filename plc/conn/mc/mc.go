package mc

import (
	"net"
)

// 프로토콜 모드
const (
	MODE_ASCII  = 1
	MODE_BINARY = 2
)

// MC Protocol 클라이언트
type MCClient struct {
	host string
	port int
	conn net.Conn
	mode int // 통신 모드 (ASCII/Binary)
}

// MC Protocol 3E 프레임 헤더
type MCHeader struct {
	SubHeader   uint16 // 0x5000 (요청) 또는 0xD000 (응답)
	NetworkNo   byte   // 네트워크 번호 (0x00)
	PCNo        byte   // PC 번호 (0xFF)
	RequestDest uint16 // 요청 대상 모듈 I/O 번호 (0x03FF)
	RequestLen  uint16 // 데이터 길이
	CPUTimer    uint16 // CPU 모니터링 타이머 (0x0010)
}

// 디바이스 코드
const (
	DEVICE_D = 0xA8 // D 레지스터
)

// 명령 코드
const (
	CMD_BATCH_READ  = 0x0401 // 일괄 읽기
	CMD_BATCH_WRITE = 0x1401 // 일괄 쓰기
)
