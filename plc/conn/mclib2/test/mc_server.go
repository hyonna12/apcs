package main

import (
	"encoding/binary"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type MCServer struct {
	memory   map[uint32]uint16 // 메모리 맵 (D레지스터)
	listener net.Listener
}

func NewMCServer() *MCServer {
	return &MCServer{
		memory: make(map[uint32]uint16),
	}
}

func (s *MCServer) Start(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.listener = listener

	log.Infof("MC Protocol Server started on port %d", port)

	// 클라이언트 연결 대기
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Errorf("Accept error: %v", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}()

	return nil
}

func (s *MCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Errorf("Read error: %v", err)
			return
		}

		log.Infof("Received request: %X", buf[:n])

		cmd := binary.BigEndian.Uint16(buf[12:14])
		subCmd := binary.BigEndian.Uint16(buf[13:15])
		log.Infof("Command: 0x%04X, SubCommand: 0x%04X", cmd, subCmd)

		switch cmd {
		case 0x0104: // 읽기 명령
			s.handleRead(conn, buf[:n])
		case 0x0114: // 쓰기 명령
			s.handleWrite(conn, buf[:n])
		default:
			log.Warnf("Unknown command: %04X", cmd)
		}
	}
}

func (s *MCServer) handleRead(conn net.Conn, req []byte) {
	// 주소와 개수 파싱
	addr := binary.LittleEndian.Uint32(req[15:18]) / 2 // 바이트 주소 -> 워드 주소
	count := binary.LittleEndian.Uint16(req[19:21])

	log.Infof("Read request: addr=D%d, count=%d", addr, count)

	// 응답 헤더 구성 (11바이트)
	resp := make([]byte, 11+count*2)
	resp[0] = 0xD0 // 응답 서브헤더
	resp[1] = 0x00

	// 데이터 읽기
	for i := uint16(0); i < count; i++ {
		value := s.memory[addr+uint32(i)]
		binary.LittleEndian.PutUint16(resp[11+i*2:], value)
	}

	conn.Write(resp)
}

func (s *MCServer) handleWrite(conn net.Conn, req []byte) {
	// 주소 파싱
	addr := binary.LittleEndian.Uint32(req[15:18]) / 2 // 바이트 주소 -> 워드 주소
	count := binary.LittleEndian.Uint16(req[19:21])

	log.Infof("Write request: addr=D%d, count=%d", addr, count)

	// 데이터 저장
	for i := uint16(0); i < count; i++ {
		value := binary.LittleEndian.Uint16(req[21+i*2:])
		s.memory[addr+uint32(i)] = value
		log.Infof("Write D%d = 0x%04X", addr+uint32(i), value)
	}

	// 응답 (11바이트)
	resp := make([]byte, 11)
	resp[0] = 0xD0 // 응답 서브헤더
	resp[1] = 0x00

	conn.Write(resp)
}

func main() {
	server := NewMCServer()
	err := server.Start(5000)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 서버 실행 유지
	select {}
}

// go test -v ./plc/conn/mclib2 -run TestMCLib2Protocol
