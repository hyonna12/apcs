package conn

import (
	"apcs_refactored/webserver"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/future-architect/go-mcprotocol/mcp"
	log "github.com/sirupsen/logrus"
)

type Task struct {
	TaskName string
	Address  string
}

var TaskList = map[string]Task{
	"Door":           {"Door", "D100"},  // (0: closed, 1: open)
	"Light":          {"Light", "D101"}, // (0: off, 1: on)
	"SetTemperature": {"SetTemperature", "D102"},
	"DetectFire":     {"DetectFire", "D103"}, // (0: no fire, 1: fire detected)
	"Weight":         {"Weight", "D104"},
	"Height":         {"Height", "D105"},
}

// 트러블 감지
func SenseTrouble(data string) {
	log.Println(data)

	/* if state.D0 == 1 {
		SenseTrouble(data)
	} */

	/* go func() {
		fmt.Println("실행")

		waiting <- "화재"
	}() */

	switch data {
	case "화재":
		log.Infof("[PLC] 화재발생")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Println("Error changing view:", err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("화재")
		//return

	case "물품 끼임":
		log.Infof("[PLC] 물품 끼임")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("물품 끼임")
		//return

	case "물품 낙하":
		log.Infof("[PLC] 물품 낙하")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("물품 낙하")

		//return

	case "이물질 감지":
		log.Infof("[PLC] 이물질 감지")
		err := webserver.ChangeKioskView("/error/trouble")
		if err != nil {
			log.Error(err)
		}
		// TODO - 사업자에게 알림
		webserver.SendEvent("이물질 감지")

		//return

	}
}

var mcClient mcp.Client

func InitConnPlc() {
	var err error

	mcClient, err = mcp.New3EClient("localhost", 6060, mcp.NewLocalStation())
	if err != nil {
		panic(err)
	}

	fmt.Println("Start connection")

	err = read("Door")
	if err != nil {
		fmt.Printf("Error reading Door: %v\n", err)
	}

	err = read("Weight")
	if err != nil {
		fmt.Printf("Error reading Weight: %v\n", err)
	}

	err = write("Door", 1)
	if err != nil {
		fmt.Printf("Error executing Door: %v\n", err)
	}

	err = read("Door")
	if err != nil {
		fmt.Printf("Error reading Door: %v\n", err)
	}
}

func parseAddress(address string) (string, int64, error) {
	deviceName := string(address[0])
	offset, err := strconv.ParseInt(address[1:], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid address format: %s", address)
	}
	return deviceName, offset, nil
}

func read(taskName string) error {
	task, exists := TaskList[taskName]
	if !exists {
		return fmt.Errorf("task %s not found", taskName)
	}
	fmt.Printf("Start Task: %s\n", taskName)

	deviceName, offset, err := parseAddress(task.Address)
	if err != nil {
		return err
	}

	response, err := mcClient.Read(deviceName, offset, 1)
	if err != nil {
		return fmt.Errorf("error reading from address %s: %v", task.Address, err)
	}

	if len(response) < 2 {
		return fmt.Errorf("invalid response length")
	}

	value := binary.LittleEndian.Uint16(response[11:13])
	//value := binary.LittleEndian.Uint16(response)
	fmt.Printf("Read value from address %s[%s]: %v\n", taskName, task.Address, value)
	return nil
}

func write(taskName string, value uint16) error {
	task, exists := TaskList[taskName]
	if !exists {
		return fmt.Errorf("task %s not found", taskName)
	}

	fmt.Printf("Start Task: %s\n", task.TaskName)

	deviceName, offset, err := parseAddress(task.Address)
	if err != nil {
		return err
	}

	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, value)

	_, err = mcClient.Write(deviceName, offset, 1, data)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", task.TaskName, err)
	}

	// Process the response
	fmt.Printf("Successfully write %s: %d\n", task.TaskName, value)

	// Verify the write operation if needed
	response, err := mcClient.Read(deviceName, offset, 1)
	if err != nil {
		return fmt.Errorf("error verifying write for %s: %v", task.TaskName, err)
	}

	if len(response) < 2 {
		return fmt.Errorf("invalid response length during verification")
	}

	readValue := binary.LittleEndian.Uint16(response[11:13])
	//readValue := binary.LittleEndian.Uint16(response)

	if readValue != value {
		return fmt.Errorf("write verification failed for %s: expected %d, got %d", task.TaskName, value, readValue)
	}

	return nil
}
