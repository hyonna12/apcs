package mc

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// 새 클라이언트 생성
func NewMCClient(host string, port int) (*MCClient, error) {
	client := &MCClient{
		host: host,
		port: port,
	}
	return client, nil
}

// 연결
func (c *MCClient) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// 연결 종료
func (c *MCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// 24비트 값 쓰기 헬퍼 함수
func putUint24(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

// D레지스터 읽기
func (c *MCClient) ReadDRegister(startAddr uint16, count uint16) ([]byte, error) {
	// 1. 헤더 구성
	header := MCHeader{
		SubHeader:   0x5000,
		NetworkNo:   0x00,
		PCNo:        0xFF,
		RequestDest: 0x03FF,
		RequestLen:  12, // 요청 데이터 길이
		CPUTimer:    0x0010,
	}

	// 2. 요청 데이터 구성
	req := make([]byte, 21)
	binary.BigEndian.PutUint16(req[0:], header.SubHeader)
	req[2] = header.NetworkNo
	req[3] = header.PCNo
	binary.BigEndian.PutUint16(req[4:], header.RequestDest)
	binary.BigEndian.PutUint16(req[6:], header.RequestLen)
	binary.BigEndian.PutUint16(req[8:], header.CPUTimer)
	binary.BigEndian.PutUint16(req[10:], CMD_BATCH_READ)
	binary.BigEndian.PutUint16(req[12:], 0x0000) // 서브 커맨드
	req[14] = DEVICE_D                           // 디바이스 코드
	putUint24(req[15:], uint32(startAddr)*2)     // 시작 주소 (워드->바이트)
	binary.BigEndian.PutUint16(req[18:], count)  // 읽을 개수

	// 3. 요청 전송
	_, err := c.conn.Write(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// 4. 응답 수신
	resp := make([]byte, 11+count*2) // 헤더(11) + 데이터
	_, err = c.conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 5. 응답 확인
	if binary.BigEndian.Uint16(resp[0:]) != 0xD000 {
		return nil, fmt.Errorf("invalid response header")
	}

	// 6. 데이터 반환
	return resp[11:], nil // 헤더 제외한 데이터만 반환
}

// D레지스터 쓰기
func (c *MCClient) WriteDRegister(addr uint16, value uint16) error {
	// 1. 헤더 구성
	header := MCHeader{
		SubHeader:   0x5000,
		NetworkNo:   0x00,
		PCNo:        0xFF,
		RequestDest: 0x03FF,
		RequestLen:  14, // 요청 데이터 길이
		CPUTimer:    0x0010,
	}

	// 2. 요청 데이터 구성
	req := make([]byte, 23)
	binary.BigEndian.PutUint16(req[0:], header.SubHeader)
	req[2] = header.NetworkNo
	req[3] = header.PCNo
	binary.BigEndian.PutUint16(req[4:], header.RequestDest)
	binary.BigEndian.PutUint16(req[6:], header.RequestLen)
	binary.BigEndian.PutUint16(req[8:], header.CPUTimer)
	binary.BigEndian.PutUint16(req[10:], CMD_BATCH_WRITE)
	binary.BigEndian.PutUint16(req[12:], 0x0000) // 서브 커맨드
	req[14] = DEVICE_D                           // 디바이스 코드
	putUint24(req[15:], uint32(addr)*2)          // 주소 (워드->바이트)
	binary.BigEndian.PutUint16(req[18:], 1)      // 쓸 개수
	binary.BigEndian.PutUint16(req[20:], value)  // 쓸 값

	// 3. 요청 전송
	_, err := c.conn.Write(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	// 4. 응답 수신
	resp := make([]byte, 11) // 응답 헤더만
	_, err = c.conn.Read(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 5. 응답 확인
	if binary.BigEndian.Uint16(resp[0:]) != 0xD000 {
		return fmt.Errorf("invalid response header")
	}

	return nil
}
