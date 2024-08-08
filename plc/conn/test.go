package conn

import (
	"fmt"
	"net"
)

type MCHeader struct {
	ProtocolVersion byte
	AddressType     byte
	Address         uint16
	Length          uint16
}

type MCPacket struct {
	Header   MCHeader
	Data     []byte
	Checksum byte
}

func InitPlc() {
	// PLC 연결 설정
	plcAddress := "localhost:6000"
	conn, err := net.Dial("tcp", plcAddress)
	if err != nil {
		fmt.Println("Failed to connect to PLC:", err)
		return
	}
	defer conn.Close()
	fmt.Println(conn)

	// MC 프로토콜 패킷 생성
	packet := MCPacket{
		Header: MCHeader{
			ProtocolVersion: 0,
			AddressType:     0,
			Address:         1000,
			Length:          3,
		},
		Data:     []byte{1, 2, 3},
		Checksum: 0, // 체크섘 계산 로직 추가 필요
	}

	// MC 프로토콜 패킷을 PLC로 전송
	packetBytes := encodePacket(packet)
	_, err = conn.Write(packetBytes)
	if err != nil {
		fmt.Println("Failed to send packet to PLC:", err)
		return
	}

	// PLC로부터 응답 수신 및 처리
	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		fmt.Println("Failed to read response from PLC:", err)
		return
	}

	// 응답 데이터 처리
	packet, err = decodePacket(response)
	if err != nil {
		fmt.Println("Failed to decode packet:", err)
		return
	}

	fmt.Printf("Received packet: %+v\n", packet)
}

func encodePacket(packet MCPacket) []byte {
	// MC 프로토콜 패킷 인코딩 로직 구현 필요
	// 예시에서는 단순히 패킷 구조체를 바이트 슬라이스로 변환
	packetBytes := []byte{
		packet.Header.ProtocolVersion,
		packet.Header.AddressType,
		byte(packet.Header.Address >> 8),
		byte(packet.Header.Address),
		byte(packet.Header.Length >> 8),
		byte(packet.Header.Length),
	}
	packetBytes = append(packetBytes, packet.Data...)
	packetBytes = append(packetBytes, packet.Checksum)

	return packetBytes
}

func decodePacket(packetBytes []byte) (MCPacket, error) {
	// MC 프로토콜 패킷 디코딩 로직 구현 필요
	// 예시에서는 단순히 바이트 슬라이스를 패킷 구조체로 변환
	if len(packetBytes) < 8 {
		return MCPacket{}, fmt.Errorf("invalid packet length")
	}

	packet := MCPacket{
		Header: MCHeader{
			ProtocolVersion: packetBytes[0],
			AddressType:     packetBytes[1],
			Address:         uint16(packetBytes[2])<<8 | uint16(packetBytes[3]),
			Length:          uint16(packetBytes[4])<<8 | uint16(packetBytes[5]),
		},
		Data:     packetBytes[6 : len(packetBytes)-1],
		Checksum: packetBytes[len(packetBytes)-1],
	}

	return packet, nil
}