package door

import (
	"apcs_refactored/plc/conn"
	"time"

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

var (
	simulatorDelay time.Duration
)

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
func SetUpDoor(DoorType DoorType, DoorOperation DoorOperation, commandId string) error {
	log.Infof("[PLC_Door] 도어 조작: %v, %v", DoorType, DoorOperation)

	err := conn.SendDoorOperation(string(DoorType), string(DoorOperation), commandId)
	if err != nil {
		log.Errorf("[PLC_Door] 도어 조작 실패: %v", err)
		return err
	}

	return nil
}
