package door

import (
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
	// TODO - PLC 도어 조작 로직
	return nil
}
