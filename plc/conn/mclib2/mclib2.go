package mclib2

import (
	"fmt"
	"time"

	mc "github.com/wang-laoban/mcprotocol"
)

type MCLib2Client struct {
	client *mc.MitsubishiClient
	host   string
	port   int
}

// 새 클라이언트 생성
func NewMCLib2Client(host string, port int) (*MCLib2Client, error) {
	// Qna_3E 프로토콜, 10초 타임아웃으로 설정
	client, err := mc.NewMitsubishiClient(mc.Qna_3E, host, port, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create MC client: %v", err)
	}

	// 연결
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &MCLib2Client{
		client: client,
		host:   host,
		port:   port,
	}, nil
}

// D레지스터 읽기
func (c *MCLib2Client) ReadDRegister(startAddr uint16, count uint16) ([]uint16, error) {
	values := make([]uint16, count)
	for i := uint16(0); i < count; i++ {
		value, err := c.client.ReadUInt16(fmt.Sprintf("D%d", startAddr+i))
		if err != nil {
			return nil, fmt.Errorf("failed to read D%d: %v", startAddr+i, err)
		}
		values[i] = value
	}
	return values, nil
}

// D레지스터 쓰기
func (c *MCLib2Client) WriteDRegister(addr uint16, value uint16) error {
	err := c.client.WriteValue(fmt.Sprintf("D%d", addr), value)
	if err != nil {
		return fmt.Errorf("failed to write D%d: %v", addr, err)
	}
	return nil
}

// 연결 종료
func (c *MCLib2Client) Close() error {
	return c.client.Close()
}
