package conn

import (
	"fmt"

	mc "github.com/wang-laoban/mcprotocol"
)

type PLCTask struct {
	TaskName string
	Address  string
}

var list = map[string]PLCTask{
	"Door":           {"Door", "D100"},  // (0: closed, 1: open)
	"Light":          {"Light", "D101"}, // (0: off, 1: on)
	"SetTemperature": {"SetTemperature", "D102"},
	"DetectFire":     {"DetectFire", "D103"}, // (0: no fire, 1: fire detected)
	"Weight":         {"Weight", "D104"},
	"Height":         {"Height", "D105"},
}

var client *mc.MitsubishiClient

func Init() {
	var err error

	// Initialize a new Mitsubishi client for the Qna_3E protocol
	client, err = mc.NewMitsubishiClient(mc.Qna_3E, "localhost", 6060, 0)
	if err != nil {
		panic(err)
	}

	fmt.Println("Start connection")

	err = client.Connect()
	if err != nil {
		panic(err)
	}

	err = readTask("Door")
	if err != nil {
		fmt.Printf("Error reading Door: %v\n", err)
	}

	err = readTask("Weight")
	if err != nil {
		fmt.Printf("Error reading Weight: %v\n", err)
	}

	err = writeTask("Door", 1)
	if err != nil {
		fmt.Printf("Error executing Door: %v\n", err)
	}

	err = readTask("Door")
	if err != nil {
		fmt.Printf("Error reading Door: %v\n", err)
	}
}

func readTask(taskName string) error {
	task, exists := list[taskName]
	if !exists {
		return fmt.Errorf("task %s not found", taskName)
	}
	fmt.Printf("Start Task: %s\n", taskName)

	value, err := client.ReadUInt16(task.Address)
	if err != nil {
		return fmt.Errorf("error reading from address %s: %v", task.Address, err)
	}

	// Process the response
	fmt.Printf("Read value from address %s[%s]: %v\n", taskName, task.Address, value)
	return nil
}

func writeTask(taskName string, value uint16) error {
	task, exists := list[taskName]
	if !exists {
		return fmt.Errorf("task %s not found", taskName)
	}

	fmt.Printf("Start Task: %s\n", task.TaskName)

	err := client.WriteValue(task.Address, value)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", task.TaskName, err)
	}

	// Process the response
	fmt.Printf("Successfully write %s: %d\n", task.TaskName, value)

	// Verify the write operation if needed
	// For example, read back the value to confirm it was written correctly
	readValue, err := client.ReadUInt16(task.Address)

	if err != nil {
		return fmt.Errorf("error verifying write for %s: %v", task.TaskName, err)
	}
	if readValue != value {
		return fmt.Errorf("write verification failed for %s: expected %d, got %d", task.TaskName, value, readValue)
	}

	return nil
}

func CloseConnection() {
	if client != nil {
		client.Close()
	}
}

// MitsubishiClient 구조체: PLC 연결 및 통신을 관리
// Connect(), Close(), ReConnect() 메서드: PLC와의 연결을 관리
// Read() 및 Write() 메서드: PLC의 데이터를 읽고 쓰는 기본 함수
// ReadBool(), ReadInt16(), ReadUInt16() 등: 특정 데이터 타입을 읽는 편의 함수
// WriteValue(): 다양한 데이터 타입을 PLC에 쓰는 함수
// GetReadCommand_Qna_3E(), GetWriteCommand_Qna_3E() 등: 프로토콜별 명령어를 생성하는 함수
// ConvertArg_Qna_3E(), ConvertArg_A_1E(): 주소 문자열을 PLC 주소 형식으로 변환하는 함수
