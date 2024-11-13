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

			// 각 클라이언트 요청 처리
			go s.handleConnection(conn)
		}
	}()

	return nil
}

func (s *MCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		// 요청 읽기
		n, err := conn.Read(buf)
		if err != nil {
			log.Errorf("Read error: %v", err)
			return
		}

		// 명령 처리
		cmd := binary.BigEndian.Uint16(buf[10:])
		switch cmd {
		case 0x0401: // 읽기
			s.handleRead(conn, buf[:n])
		case 0x1401: // 쓰기
			s.handleWrite(conn, buf[:n])
		}
	}
}

func (s *MCServer) handleRead(conn net.Conn, req []byte) {
	// 주소와 개수 파싱
	addr := uint32(req[15])<<16 | uint32(req[16])<<8 | uint32(req[17])
	count := binary.BigEndian.Uint16(req[18:])
	addr = addr / 2 // 바이트 주소 -> 워드 주소

	// 응답 헤더 구성
	resp := make([]byte, 11+count*2)
	binary.BigEndian.PutUint16(resp[0:], 0xD000) // 응답 헤더

	// 데이터 읽기
	for i := uint16(0); i < count; i++ {
		value := s.memory[addr+uint32(i)]
		binary.BigEndian.PutUint16(resp[11+i*2:], value)
	}

	conn.Write(resp)
}

func (s *MCServer) handleWrite(conn net.Conn, req []byte) {
	// 주소 파싱
	addr := uint32(req[15])<<16 | uint32(req[16])<<8 | uint32(req[17])
	addr = addr / 2 // 바이트 주소 -> 워드 주소

	// 값 저장
	value := binary.BigEndian.Uint16(req[20:])
	s.memory[addr] = value

	// 응답
	resp := make([]byte, 11)
	binary.BigEndian.PutUint16(resp[0:], 0xD000)
	conn.Write(resp)

	log.Infof("Write D%d = 0x%04X", addr, value)
}
