package mclib

import (
	"encoding/binary"
	"fmt"

	"github.com/future-architect/go-mcprotocol/mcp"
)

// MC Protocol 클라이언트
type MCLibClient struct {
	client mcp.Client
	host   string
	port   int
}

// 새 클라이언트 생성
func NewMCLibClient(host string, port int) (*MCLibClient, error) {
	client, err := mcp.New3EClient(host, port, mcp.NewLocalStation())
	if err != nil {
		return nil, fmt.Errorf("failed to create MC client: %v", err)
	}

	return &MCLibClient{
		client: client,
		host:   host,
		port:   port,
	}, nil
}

// D레지스터 읽기
func (c *MCLibClient) ReadDRegister(startAddr uint16, count uint16) ([]byte, error) {
	data, err := c.client.Read("D", int64(startAddr), int64(count))
	if err != nil {
		return nil, fmt.Errorf("failed to read D register: %v", err)
	}
	return data[11:], nil // 헤더 제외하고 데이터만 반환
}

// D레지스터 쓰기
func (c *MCLibClient) WriteDRegister(addr uint16, value uint16) error {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, value)

	_, err := c.client.Write("D", int64(addr), 1, data)
	if err != nil {
		return fmt.Errorf("failed to write D register: %v", err)
	}
	return nil
}

// 연결 상태 확인
func (c *MCLibClient) HealthCheck() error {
	return c.client.HealthCheck()
}
