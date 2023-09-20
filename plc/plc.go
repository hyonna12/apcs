package plc

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
	"apcs_refactored/model"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/robot"
	"apcs_refactored/plc/sensor"
	"apcs_refactored/plc/trayBuffer"
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	Robot = iota
	Slot
	Tray
	Table
	Gate
)

var (
	msgNode        *messenger.Node
	simulatorDelay time.Duration
	// TODO - temp - 키오스크 임시 물건 꺼내기 버튼 시뮬레이션 용
	IsItemOnTable = false
	// TODO - temp - 빈 트레이 감지 시뮬레이션 용
	TrayIdOnTable sql.NullInt64
)

type ItemDimension struct {
	Height      int
	Width       int
	Weight      int
	TrackingNum int
}

func StartPlcClient(n *messenger.Node) {
	msgNode = n

	// 시뮬레이터 딜레이 설정
	simulatorDelay = time.Duration(config.Config.Plc.Simulation.Delay)

	robot.InitRobots()

	// TODO - 삭제
	go msgNode.ListenMessages(
		func(m *messenger.Message) bool {
			log.Debugf("[%v] Received Message: %v", msgNode.Name, m)
			// TODO - PLC에 명령을 전달하는데 성공하면 true
			return true
		},
	)

	// Job을 로봇에 배정하는 고루틴
	go robot.DistributeJob()

}

// SetUpDoor
//
// 도어 조작.
//
// - door.DoorType: 조작할 도어
// - door.DoorOperation: 조작 명령
func SetUpDoor(doorType door.DoorType, doorOperation door.DoorOperation) error {
	err := door.SetUpDoor(doorType, doorOperation)
	if err != nil {
		return err
	}
	return nil
}

// SetUpTrayBuffer
//
// 트레이 버퍼 조작.
//
// - tray.BufferOperation: 조작 명령
func SetUpTrayBuffer(bufferOperation trayBuffer.BufferOperation) error {
	err := trayBuffer.SetUpTrayBuffer(bufferOperation)
	if err != nil {
		log.Error("버퍼 조작 에러", err)

		return err
	}
	return nil
}

// SenseTableForItem
//
// 테이블에 물품이 있는지 감지.
// 있으면 true, 없으면 false 반환.
func SenseTableForItem() (bool, error) {
	IsItemOnTable, err := sensor.SenseTableForItem()
	if err != nil {
		return IsItemOnTable, err
	}

	return IsItemOnTable, nil
}

// ServeEmptyTrayToTable
//
// 테이블에 빈 트레이를 서빙하도록 함.
// 본 함수 호출 전 SenseTableForEmptyTray 함수 호출하여 빈 트레이 여부 확인 후 로직 분기 필요함.
//
// - slot: 빈 트레이를 가져올 슬롯
func ServeEmptyTrayToTable(slot model.Slot) error {
	log.Infof("[PLC] 테이블에 빈 트레이 서빙")

	if err := robot.JobServeEmptyTrayToTable(slot); err != nil {
		return err
	}
	TrayIdOnTable = slot.TrayId

	return nil
}

// StandbyRobotAtTable
//
// 여유 로봇 한 대를 테이블 앞에 대기시킴.
func StandbyRobotAtTable() error {
	if err := robot.JobWaitAtTable(); err != nil {
		return err
	}

	return nil
}

func DismissRobotAtTable() error {
	// TODO - 로봇 대기 상태 해제
	if err := robot.JobDismiss(); err != nil {
		return err
	}

	return nil
}

// SenseItemInfo
//
// 테이블에 올려진 물품 크기/무게 계측.
func SenseItemInfo() (ItemDimension, error) {
	itemDimension, err := sensor.SenseItemInfo()
	ItemDimension := ItemDimension{Height: itemDimension.Height, Weight: itemDimension.Weight, Width: itemDimension.Width}

	if err != nil {
		return ItemDimension, err
	}

	return ItemDimension, nil
}

// MoveTray
//
// 트레이를 한 슬롯에서 다른 슬롯으로 이동(정리).
//
// - from: 트레이를 꺼낼 슬롯
// - to: 트레이를 넣을 슬롯
func MoveTray(from, to model.Slot) error {
	log.Infof("[PLC] 트레이 이동: fromSlotId=%v -> toSlotId=%v", from.SlotId, to.SlotId)

	if err := robot.JobMoveTray(from, to); err != nil {
		return err
	}

	return nil
}

// RetrieveEmptyTrayFromTable
//
// 테이블의 빈 트레이 회수.
// 회수한 트레이의 id를 반환
//
// - slot: 빈 트레이를 격납할 슬롯	***************
func RetrieveEmptyTrayFromTable(slot model.Slot) (int64, error) {
	log.Infof("[PLC] 테이블 빈 트레이 회수. slotId=%v", slot.SlotId)

	if err := robot.JobRetrieveEmptyTrayFromTable(slot); err != nil {
		return 0, err
	}

	trayId := TrayIdOnTable.Int64
	//TrayIdOnTable = sql.NullInt64{Valid: false} // set null

	return trayId, nil
}

// InputItem
//
// 물품 수납. 트레이 id 반환
//
// - slot: 물품을 수납할 슬롯
func InputItem(slot model.Slot) (int64, error) {
	log.Infof("[PLC] 물품 수납. slotId=%v", slot.SlotId)

	if err := robot.JobInputItem(slot); err != nil {
		return 0, err
	}

	trayId := TrayIdOnTable.Int64
	//trayIdOnTable := sql.NullInt64{Valid: false} // set null

	return trayId, nil
}

// OutputItem
//
// 물품 불출.
//
// - slot: 물품을 꺼내올 슬롯	***************
func OutputItem(slot model.Slot) error {
	log.Infof("[PLC] 물품 불출 시작. 꺼내올 슬롯 id=%v", slot.SlotId)

	if err := robot.JobOutputItem(slot); err != nil {
		return err
	}

	// TODO - temp - 시뮬레이션용
	IsItemOnTable = true

	TrayIdOnTable = slot.TrayId

	log.Infof("[PLC] 물품 불출 완료. 꺼내온 슬롯 id=%v", slot.SlotId)
	return nil
}
