package door

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type DoorType string
type DoorOperation string

const (
	DoorTypeFront DoorType = "DoorTypeFront"
	DoorTypeBack  DoorType = "DoorTypeBack"

	DoorOperationOpen  DoorOperation = "DoorOperationOpen"
	DoorOperationClose DoorOperation = "DoorOperationClose"
)

type DoorState struct {
	DoorType DoorType `json:"type"`
	State    string   `json:"state"`
}

type Door struct {
	DoorType      DoorType      `json:"DoorType"`
	DoorOperation DoorOperation `json:"DoorOperation"`
}

// SetUpDoor
//
// 도어 조작.
//
// - door.DoorType: 조작할 도어
// - door.DoorOperation: 조작 명령
func SetUpDoor(DoorType DoorType, DoorOperation DoorOperation) error {
	log.Infof("[PLC_Door] 도어 조작: %v, %v", DoorType, DoorOperation)
	// PLC 도어 조작
	data := Door{DoorType, DoorOperation}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	_, err := http.Post("http://localhost:8000/setup/door", "application/json", buff)
	if err != nil {
		return err
	}

	return nil
}

// GetDoorState
//
// 도어 상태 조회
//
// - door.DoorType: 조작할 도어
// - door.DoorOperation: 조작 명령
func GetDoorState() ([]DoorState, error) {
	log.Infof("[PLC_Door] 도어 상태 조회")
	// PLC 도어 조작 로직
	var doorState []DoorState
	resp, err := http.Get("http://localhost:8000/door")
	if err != nil {
		return doorState, err
	}

	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return doorState, err
	}
	json.Unmarshal(respData, &doorState)

	log.Infof("[PLC_Door] 도어 상태: %v", doorState)

	return doorState, err
}
