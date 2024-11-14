package mc

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// NewMCClient 함수 수정
func NewMCClient(host string, port int, mode int) (*MCClient, error) {
	if mode != MODE_ASCII && mode != MODE_BINARY {
		return nil, fmt.Errorf("잘못된 프로토콜 모드: %d", mode)
	}

	client := &MCClient{
		host: host,
		port: port,
		mode: mode,
	}
	return client, nil
}

// 연결
func (c *MCClient) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("PLC 연결 실패: %v", err)
	}
	// 읽기/쓰기 타임아웃 설정
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	c.conn = conn
	log.Infof("PLC 연결 성공: %s", addr)
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

// PLCValue 구조체 정의
type PLCValue struct {
	Raw    []byte // 원본 데이터
	Length int    // 데이터 길이
}

// 값 변환을 위한 메서드들
func (v *PLCValue) ToUint16() (uint16, error) {
	if v.Length < 2 {
		return 0, fmt.Errorf("데이터 길이 부족 (uint16): %d", v.Length)
	}
	return binary.BigEndian.Uint16(v.Raw), nil
}

func (v *PLCValue) ToInt16() (int16, error) {
	val, err := v.ToUint16()
	if err != nil {
		return 0, err
	}
	return int16(val), nil
}

func (v *PLCValue) ToBinary() string {
	val, err := v.ToUint16()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("%016b", val)
}

func (v *PLCValue) ToHex() string {
	val, err := v.ToUint16()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("0x%04X", val)
}

func (v *PLCValue) GetBit(pos uint) (bool, error) {
	if pos > 15 {
		return false, fmt.Errorf("잘못된 비트 위치: %d", pos)
	}
	val, err := v.ToUint16()
	if err != nil {
		return false, err
	}
	return (val & (1 << pos)) != 0, nil
}

// 10진수 문자열로 변환
func (v *PLCValue) ToDecimal() string {
	val, err := v.ToUint16()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("%d", val)
}

// 부호있는 10진수 문자열로 변환
func (v *PLCValue) ToSignedDecimal() string {
	val, err := v.ToInt16()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("%d", val)
}

// 워드 값을 가져오는 메서드들 추가
func (v *PLCValue) ToWord() string {
	val, err := v.ToUint16()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	// 워드 값을 10진수, 16진수, 2진수로 표현
	return fmt.Sprintf("W%d (0x%04X, %016b)", val, val, val)
}

// 워드 단위로 모든 형식 출력
func (v *PLCValue) ToString() string {
	uint16Val, _ := v.ToUint16()
	int16Val, _ := v.ToInt16()
	return fmt.Sprintf("WORD: W%d, DEC: %d, HEX: %s, BIN: %s, SIGNED: %d",
		uint16Val,
		uint16Val,
		v.ToHex(),
		v.ToBinary(),
		int16Val)
}

// ReadDRegister 함수 수정
func (c *MCClient) ReadDRegister(startAddr uint16, count uint16) ([]*PLCValue, error) {
	data, err := c.readDRegisterRaw(startAddr, count)
	if err != nil {
		return nil, err
	}

	// 데이터를 PLCValue 배열로 변환
	values := make([]*PLCValue, count)
	for i := uint16(0); i < count; i++ {
		if len(data) < int((i+1)*2) {
			return nil, fmt.Errorf("데이터 길이 부족: %d", len(data))
		}
		values[i] = &PLCValue{
			Raw:    data[i*2 : (i+1)*2],
			Length: 2,
		}
	}

	return values, nil
}

func (c *MCClient) readDRegisterASCII(startAddr uint16, count uint16) ([]byte, error) {
	// 1. 요청 데이터 구성
	reqStr := fmt.Sprintf("5000FF03FF000C00100401000001%04X%04X", startAddr, count)
	req := []byte(reqStr)

	log.Debugf("=== ASCII 읽기 요청 시작 ===")
	log.Debugf("요청 주소: D%d, 개수: %d", startAddr, count)
	log.Debugf("요청 데이터: %s", reqStr)
	log.Debugf("요청 데이터 분석:")
	log.Debugf("- 헤더: %s", reqStr[:4])
	log.Debugf("- 네트워크/PC: %s", reqStr[4:8])
	log.Debugf("- 데이터 길이: %s", reqStr[8:12])
	log.Debugf("- 명령 코드: %s", reqStr[12:16])
	log.Debugf("- 주소/개수: %s", reqStr[16:])

	// 2. 요청 전송
	_, err := c.conn.Write(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %v", err)
	}
	log.Debug("요청 전송 완료")

	// 3. 응답 수신
	resp := make([]byte, 1024)
	n, err := c.conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("응답 수신 실패: %v", err)
	}
	resp = resp[:n]

	log.Debugf("=== ASCII 응답 수신 ===")
	log.Debugf("응답 길이: %d bytes", n)
	log.Debugf("응답 전체: %s", string(resp))

	// 4. 응답 데이터 분석
	if n >= 4 {
		log.Debugf("응답 헤더: %s", string(resp[:4]))
	}
	if n >= 14 {
		log.Debugf("응답 상세:")
		log.Debugf("- 네트워크/PC 번호: %s", string(resp[4:8]))
		log.Debugf("- 데이터 길이: %s", string(resp[8:12]))
		log.Debugf("- 응답 코드: %s", string(resp[12:14]))
	}
	if n >= 36 {
		log.Debugf("- 데이터 영역: %s", string(resp[14:]))
		log.Debugf("- 최종 데이터: %s", string(resp[n-4:n]))
	}

	// 5. 응답 헤더 검증
	if !((len(resp) >= 4 && string(resp[:4]) == "D000") ||
		(len(resp) >= 4 && string(resp[:4]) == "50B0")) {
		return nil, fmt.Errorf("잘못된 응답 헤더: %s", string(resp[:4]))
	}

	// 6. 50B0 응답 처리
	if string(resp[:4]) == "50B0" {
		log.Debug("50B0 응답 수신 - 기본값 반환")
		data := make([]byte, count*2)
		return data, nil
	}

	// 7. 데이터 변환
	data := make([]byte, count*2)
	if n < 36 {
		return nil, fmt.Errorf("응답 데이터 길이 부족: %d", n)
	}

	// 마지막 4바이트 처리
	hexStr := string(resp[n-4 : n])
	log.Debugf("변환할 16진수 문자열: %s", hexStr)
	
	val, err := strconv.ParseUint(hexStr, 16, 32)  // 32비트로 변경
	if err != nil {
		log.Debugf("파싱 실패한 문자열: %q", hexStr)
		return nil, fmt.Errorf("데이터 변환 실패: %v", err)
	}
	log.Debugf("변환된 값: 0x%08X", val)

	// 8. 결과 저장 (리틀 엔디안)
	if count == 2 {
		// 2워드인 경우 32비트 값으로 처리
		binary.LittleEndian.PutUint32(data, uint32(val))
	} else {
		// 1워드인 경우 16비트 값으로 처리
		binary.BigEndian.PutUint16(data, uint16(val))
	}
	
	log.Debugf("=== 읽기 완료 ===")
	return data, nil
}

func (c *MCClient) WriteDRegister(addr uint16, value uint16) error {
	if c.mode == MODE_ASCII {
		return c.writeDRegisterASCII(addr, value)
	}
	return c.writeDRegisterBinary(addr, value)
}

func (c *MCClient) writeDRegisterASCII(addr uint16, value uint16) error {
	// ASCII 모드 요청 데이터 구성
	reqStr := fmt.Sprintf("5000FF03FF000E00101401000000%04X0001%04X", addr, value)
	req := []byte(reqStr)

	log.Debugf("ASCII 쓰기 요청: %s", reqStr)

	// 요청 전송
	_, err := c.conn.Write(req)
	if err != nil {
		return fmt.Errorf("요청 전송 실패: %v", err)
	}

	// 응답 수신 버퍼 크기 증가
	resp := make([]byte, 1024)
	n, err := c.conn.Read(resp)
	if err != nil {
		return fmt.Errorf("응답 수신 실패: %v", err)
	}
	resp = resp[:n]

	log.Debugf("ASCII 쓰기 응답: %s (길이: %d)", string(resp), n)

	// 응답 헤더 검증 (D000으로 시작하거나 50B0으로 시작)
	if !((len(resp) >= 4 && string(resp[:4]) == "D000") ||
		(len(resp) >= 4 && string(resp[:4]) == "50B0")) {
		return fmt.Errorf("잘못된 응답 헤더: %s", string(resp[:4]))
	}

	// 50B0 응답은 성공 응답의 다른 형태
	if string(resp[:4]) == "50B0" {
		log.Debug("50B0 응답 수신 - 쓰기 성공")
		return nil
	}

	// 응답 코드 확인 (마지막 2바이트가 0000이어야 함)
	if n >= 6 {
		respCode := string(resp[n-4 : n])
		if respCode != "0000" {
			return fmt.Errorf("잘못된 응답 코드: %s", respCode)
		}
	}

	return nil
}

// Binary 모드 읽기 구현
func (c *MCClient) readDRegisterBinary(startAddr uint16, count uint16) ([]byte, error) {
	// 1. 요청 프레임 구성
	req := make([]byte, 21)
	binary.BigEndian.PutUint16(req[0:], 0x5000)  // 서브헤더
	req[2] = 0x00                                // 네트워크 번호
	req[3] = 0xFF                                // PC 번호
	binary.BigEndian.PutUint16(req[4:], 0x03FF)  // 요청 대상
	binary.BigEndian.PutUint16(req[6:], 12)      // 요청 데이터 길이
	binary.BigEndian.PutUint16(req[8:], 0x0010)  // CPU 감시 타이머
	binary.BigEndian.PutUint16(req[10:], 0x0401) // 명령 (일괄 읽기)
	binary.BigEndian.PutUint16(req[12:], 0x0000) // 서브 명령
	req[14] = 0xA8                               // 디바이스 코드 (D)
	putUint24(req[15:], uint32(startAddr))       // 시작 주소
	binary.BigEndian.PutUint16(req[18:], count)  // 읽을 개수

	log.Debugf("Binary 요청: %X", req)

	// 2. 요청 전송
	_, err := c.conn.Write(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %v", err)
	}

	// 3. 응답 수신 (버퍼 크기 증가)
	resp := make([]byte, 1024)
	n, err := c.conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("응답 수신 실패: %v", err)
	}
	resp = resp[:n]

	log.Debugf("Binary 응답 (길이: %d): %X", n, resp)

	// 4. 응답 형식 확인
	if n < 11 {
		return nil, fmt.Errorf("응답 길이 부족: %d", n)
	}

	// ASCII 응답인 경우 처리
	if resp[0] == '5' || resp[0] == 'D' {
		log.Debug("ASCII 응답 감지, ASCII 모드로 처리")
		data := make([]byte, count*2)

		// ASCII 응답 파싱
		for i := 0; i < int(count); i++ {
			startIdx := 16 + i*4
			if startIdx+4 > len(resp) {
				return nil, fmt.Errorf("데이터 길이 부족 (위치 %d)", i)
			}

			hexStr := string(resp[startIdx : startIdx+4])
			val, err := strconv.ParseUint(hexStr, 16, 16)
			if err != nil {
				return nil, fmt.Errorf("데이터 변환 실패 (위치 %d): %v", i, err)
			}

			binary.BigEndian.PutUint16(data[i*2:], uint16(val))
		}
		return data, nil
	}

	// 바이너리 응답 처리
	if binary.BigEndian.Uint16(resp[0:]) != 0xD000 {
		return nil, fmt.Errorf("잘못된 응답 헤더: %X", resp[0:2])
	}

	// 5. 데이터 반환
	return resp[11 : 11+count*2], nil
}

// Binary 모드 쓰기 구현
func (c *MCClient) writeDRegisterBinary(addr uint16, value uint16) error {
	// 1. 요청 프레임 구성
	req := make([]byte, 23)
	binary.BigEndian.PutUint16(req[0:], 0x5000)  // 서브헤더
	req[2] = 0x00                                // 네트워크 번호
	req[3] = 0xFF                                // PC 번호
	binary.BigEndian.PutUint16(req[4:], 0x03FF)  // 요청 대상
	binary.BigEndian.PutUint16(req[6:], 14)      // 요청 데이터 길이
	binary.BigEndian.PutUint16(req[8:], 0x0010)  // CPU 감시 타이머
	binary.BigEndian.PutUint16(req[10:], 0x1401) // 명령 (일괄 쓰기)
	binary.BigEndian.PutUint16(req[12:], 0x0000) // 서브 명령
	req[14] = 0xA8                               // 디바이스 코드 (D)
	putUint24(req[15:], uint32(addr))            // 주소
	binary.BigEndian.PutUint16(req[18:], 1)      // 쓸 개수
	binary.BigEndian.PutUint16(req[20:], value)  // 쓸 값

	log.Debugf("Binary 쓰기 요청: %X", req)

	// 2. 요청 전송
	_, err := c.conn.Write(req)
	if err != nil {
		return fmt.Errorf("요청 전송 실패: %v", err)
	}

	// 3. 응답 수신 (버퍼 크기 증가)
	resp := make([]byte, 1024)
	n, err := c.conn.Read(resp)
	if err != nil {
		return fmt.Errorf("응답 수신 실패: %v", err)
	}
	resp = resp[:n]

	log.Debugf("Binary 쓰기 응답 (길이: %d): %X", n, resp)

	// ASCII 응답인 경우 처리
	if resp[0] == '5' || resp[0] == 'D' {
		if string(resp[:4]) != "D000" {
			return fmt.Errorf("잘못된 ASCII 응답 헤더: %s", string(resp[:4]))
		}
		return nil
	}

	// 바이너리 응답 처리
	if binary.BigEndian.Uint16(resp[0:]) != 0xD000 {
		return fmt.Errorf("잘못된 응답 헤더: %X", resp[0:2])
	}

	return nil
}

// 2워드 쓰기 함수 수정 (데이터 타입 고려)
func (c *MCClient) WriteDRegister2Word(addr uint16, value1 uint16, value2 uint16) error {
	log.Debugf("=== 2워드 쓰기 시작 ===")
	log.Debugf("대상: D%d-D%d", addr, addr+1)
	log.Debugf("값: D%d=0x%04X, D%d=0x%04X", addr, value1, addr+1, value2)

	// ASCII 모드 요청 데이터 구성
	// 데이터 길이를 0x14로 수정 (2워드 데이터 + 헤더)
	reqStr := fmt.Sprintf("5000FF03FF001400101401000001%04X0002%04X%04X",
		addr,   // 시작 주소
		value1, // D5500 값
		value2) // D5501 값

	log.Debugf("요청 데이터: %s", reqStr)
	log.Debugf("요청 분석:")
	log.Debugf("- 헤더: %s", reqStr[:4])
	log.Debugf("- 네트워크/PC: %s", reqStr[4:8])
	log.Debugf("- 데이터 길이: %s", reqStr[8:12])
	log.Debugf("- 명령 코드: %s", reqStr[12:16])
	log.Debugf("- 주소: %s", reqStr[16:20])
	log.Debugf("- 개수: %s", reqStr[20:24])
	log.Debugf("- 데이터1: %s", reqStr[24:28])
	log.Debugf("- 데이터2: %s", reqStr[28:])

	req := []byte(reqStr)

	// 요청 전송
	_, err := c.conn.Write(req)
	if err != nil {
		return fmt.Errorf("요청 전송 실패: %v", err)
	}

	// 응답 수신
	resp := make([]byte, 1024)
	n, err := c.conn.Read(resp)
	if err != nil {
		return fmt.Errorf("응답 수신 실패: %v", err)
	}
	resp = resp[:n]
	log.Debugf("응답: %s (길이: %d)", string(resp), n)

	// 응답 검증
	if !((len(resp) >= 4 && string(resp[:4]) == "D000") ||
		(len(resp) >= 4 && string(resp[:4]) == "50B0")) {
		return fmt.Errorf("잘못된 응답 헤더: %s", string(resp[:4]))
	}

	// 쓰기 확인
	time.Sleep(200 * time.Millisecond) // 지연 시간 증가
	values, err := c.ReadDRegister(addr, 2)
	if err != nil {
		return fmt.Errorf("쓰기 확인 실패: %v", err)
	}

	if len(values) >= 2 {
		val1, _ := values[0].ToUint16()
		val2, _ := values[1].ToUint16()

		log.Debugf("확인된 값: D%d=0x%04X, D%d=0x%04X",
			addr, val1, addr+1, val2)

		if val1 != value1 || val2 != value2 {
			return fmt.Errorf("쓰기 ��� 불일치: got [0x%04X,0x%04X], want [0x%04X,0x%04X]",
				val1, val2, value1, value2)
		}
	}

	return nil
}

// 32비트 값 쓰기 함수 수정
func (c *MCClient) WriteDRegister32(addr uint16, value uint32) error {
	// 빅 엔디안으로 워드 분리 (PLC 메모리 구조에 맞춤)
	highWord := uint16(value >> 16)   // 상위 워드 (D5500)
	lowWord := uint16(value & 0xFFFF) // 하위 워드 (D5501)

	log.Debugf("32비트 값 분할:")
	log.Debugf("원본 값: 0x%08X (%d)", value, value)
	log.Debugf("- D%d: 0x%04X (상위 워드)", addr, highWord)
	log.Debugf("- D%d: 0x%04X (하위 워드)", addr+1, lowWord)

	// 순서대로 쓰기
	return c.WriteDRegister2Word(addr, highWord, lowWord)
}

// 32비트(2워드) 값을 처리하기 위한 메서드 추가
func (v *PLCValue) ToUint32() (uint32, error) {
	if v.Length < 4 {
		return 0, fmt.Errorf("데이터 길이 부족 (uint32): %d", v.Length)
	}
	return binary.BigEndian.Uint32(v.Raw), nil
}

func (v *PLCValue) ToInt32() (int32, error) {
	val, err := v.ToUint32()
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// 2워드 읽기 함수 수정 (리틀 엔디안 처리)
func (c *MCClient) ReadDRegister2Word(addr uint16) (*PLCValue, error) {
	log.Debugf("=== 2워드 읽기 시작 (D%d-D%d) ===", addr, addr+1)

	// 2워드(4바이트) 읽기
	data, err := c.readDRegisterRaw(addr, 2)
	if err != nil {
		log.Errorf("2워드 읽기 실패: %v", err)
		return nil, err
	}

	// 데이터 길이 확인
	if len(data) < 4 {
		log.Errorf("데이터 길이 부족: %d bytes (4 bytes 필요)", len(data))
		return nil, fmt.Errorf("데이터 길이 부족 (2워드): %d", len(data))
	}

	// 원본 데이터 로깅
	log.Debugf("수신된 원본 데이터: %X", data)

	// PLC 메모리 구조 (1234567 = 0x0012D687)
	// D5500: 0xD687 (하위 워드)
	// D5501: 0x0012 (상위 워드)
	log.Debugf("바이트별 분석:")
	log.Debugf("- D%d (하위 워드): [%02X %02X]", addr, data[0], data[1])
	log.Debugf("- D%d (상위 워드): [%02X %02X]", addr+1, data[2], data[3])

	// 데이터를 하나의 PLCValue로 변환
	value := &PLCValue{
		Raw:    make([]byte, 4),
		Length: 4,
	}

	// PLC의 워드 순서에 맞게 조정
	// 예: 1234567 (0x0012D687)
	value.Raw[0] = data[2]  // 상위 워드 상위 바이트 (0x00)
	value.Raw[1] = data[3]  // 상위 워드 하위 바이트 (0x12)
	value.Raw[2] = data[0]  // 하위 워드 상위 바이트 (0xD6)
	value.Raw[3] = data[1]  // 하위 워드 하위 바이트 (0x87)

	// 변환된 값 로깅
	uint32Val, _ := value.ToUint32()
	log.Debugf("변환된 32비트 값: 0x%08X (%d)", uint32Val, uint32Val)
	log.Debugf("개별 워드 값:")
	log.Debugf("- D%d (하위): 0x%04X", addr, binary.BigEndian.Uint16(data[0:2]))
	log.Debugf("- D%d (상위): 0x%04X", addr+1, binary.BigEndian.Uint16(data[2:4]))

	log.Debug("=== 2워드 읽기 완료 ===")
	return value, nil
}

// PLCValue에 32비트 값 처리 메서드 추가
func (v *PLCValue) ToDoubleWord() string {
	val, err := v.ToUint32()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("DW%d (0x%08X, %032b)", val, val, val)
}

// 32비트 값의 모든 형식 출력
func (v *PLCValue) ToString32() string {
	uint32Val, err := v.ToUint32()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	int32Val := int32(uint32Val)
	return fmt.Sprintf("DWORD: DW%d, DEC: %d, HEX: 0x%08X, BIN: %032b, SIGNED: %d",
		uint32Val,
		uint32Val,
		uint32Val,
		uint32Val,
		int32Val)
}

// MC 프로토콜 응답 구조체
type MCResponse struct {
	SubHeader   string // D000
	NetworkNo   byte   // FF
	PCNo        byte   // 03
	RequestDest uint16 // FF00
	DataLen     uint16 // 000C
	EndCode     uint16 // 0000
	CPUTimer    uint16 // 0016
	Command     uint16 // 4A00
	SubCommand  uint16 // 0000
	Data        []byte // 실제 데이터
}

// 응답 데이터 파싱 함수
func parseMCResponse(resp []byte) (*MCResponse, error) {
	if len(resp) < 36 {
		return nil, fmt.Errorf("응답 길이 부족: %d", len(resp))
	}

	// ASCII 형식 응답 파싱
	mcResp := &MCResponse{
		SubHeader:   string(resp[0:4]),                                   // D000
		NetworkNo:   resp[4],                                             // FF
		PCNo:        resp[5],                                             // 03
		RequestDest: binary.BigEndian.Uint16([]byte{resp[6], resp[7]}),   // FF00
		DataLen:     binary.BigEndian.Uint16([]byte{resp[8], resp[9]}),   // 000C
		EndCode:     binary.BigEndian.Uint16([]byte{resp[10], resp[11]}), // 0000
		CPUTimer:    binary.BigEndian.Uint16([]byte{resp[12], resp[13]}), // 0016
		Command:     binary.BigEndian.Uint16([]byte{resp[14], resp[15]}), // 4A00
		SubCommand:  binary.BigEndian.Uint16([]byte{resp[16], resp[17]}), // 0000
		Data:        resp[len(resp)-4:],                                  // 마지막 4바이트가 데이터
	}

	// 응답 헤더 검증
	if mcResp.SubHeader != "D000" && mcResp.SubHeader != "50B0" {
		return nil, fmt.Errorf("잘못된 응답 헤더: %s", mcResp.SubHeader)
	}

	// 에러 코드 확인
	if mcResp.EndCode != 0 {
		return nil, fmt.Errorf("PLC 에러 응답: %04X", mcResp.EndCode)
	}

	return mcResp, nil
}

// CheckConnection 함수 수정
func (c *MCClient) CheckConnection() error {
	if c.conn == nil {
		return fmt.Errorf("PLC 연결이 설정되지 않음")
	}

	// 간단한 읽기 테스트
	values, err := c.ReadDRegister(5000, 1)
	if err != nil {
		return fmt.Errorf("PLC 연결 테스트 실패: %v", err)
	}

	if len(values) > 0 {
		log.Infof("PLC 연결 테스트 성공: %s", values[0].ToWord())
		return nil
	}

	return fmt.Errorf("응답 데이터 없음")
}

// readDRegisterRaw는 모드에 따라 적절한 읽기 함수를 호출합니다
func (c *MCClient) readDRegisterRaw(startAddr uint16, count uint16) ([]byte, error) {
	log.Debugf("=== 레지스터 읽기 시작 ===")
	log.Debugf("주소: D%d, 개수: %d 워드", startAddr, count)
	log.Debugf("모드: %s", map[int]string{MODE_ASCII: "ASCII", MODE_BINARY: "Binary"}[c.mode])

	var data []byte
	var err error

	if c.mode == MODE_ASCII {
		data, err = c.readDRegisterASCII(startAddr, count)
	} else {
		data, err = c.readDRegisterBinary(startAddr, count)
	}

	if err != nil {
		log.Errorf("읽기 실패: %v", err)
		return nil, err
	}

	log.Debugf("읽기 결과: %X (길이: %d bytes)", data, len(data))
	log.Debug("=== 레지스터 읽기 완료 ===")

	return data, nil
}
