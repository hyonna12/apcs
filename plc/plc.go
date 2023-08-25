package plc

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
	"apcs_refactored/model"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/robot"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	Robot = "Robot"
	Slot  = "Slot"
	Tray  = "Tray"
	Table = "Table"
	Gate  = "Gate"
)

var (
	msgNode        *messenger.Node
	simulatorDelay time.Duration
)

type ItemDimension struct {
	Height      int
	Width       int
	Length      int
	TrackingNum int
}

func StartPlcClient(n *messenger.Node) {
	msgNode = n

	// 시뮬레이터 딜레이 설정
	simulatorDelay = time.Duration(config.Config.Plc.Simulation.Delay)

	robot.InitRobots()

	go msgNode.ListenMessages(
		func(m *messenger.Message) bool {
			log.Debugf("[%v] Received Message: %v", msgNode.Name, m)
			// TODO - PLC에 명령을 전달하는데 성공하면 true
			return true
		},
	)

	// Job을 로봇에 배정
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

// SenseTableForEmptyTray
//
// 테이블에 빈 트레이가 있는지 감지.
// 있으면 true, 없으면 false 반환.
func SenseTableForEmptyTray() (bool, error) {
	log.Infof("[PLC_Sensor] 테이블 빈 트레이 감지")
	// TODO - PLC 센서 빈 트레이 감지 로직
	return false, nil
}

// SenseTableForItem
//
// 테이블에 물품이 있는지 감지.
// 있으면 true, 없으면 false 반환.
func SenseTableForItem() (bool, error) {
	log.Infof("[PLC_Sensor] 테이블 물품 존재 여부 감지")
	// TODO - PLC 센서 물품 존재 여부 감지
	return false, nil
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
	log.Infof("[PLC] 크기/무게 측정")

	// TODO - temp
	itemDimension := &ItemDimension{
		Height:      rand.Intn(40),
		Width:       rand.Intn(40),
		Length:      rand.Intn(40),
		TrackingNum: rand.Intn(40),
	}

	return *itemDimension, nil
}

// MoveTray
//
// 트레이를 한 슬롯에서 다른 슬롯으로 이동(정리).
//
// - from: 트레이를 꺼낼 슬롯
// - to: 트레이를 넣을 슬롯
func MoveTray(from, to model.Slot) error {
	log.Infof("[PLC] 트레이 이동: %v -> %v", from, to)
	// TODO -

	return nil
}

// RetrieveEmptyTrayFromTable
//
// 테이블의 빈 트레이 회수.
//
// - slot: 빈 트레이를 격납할 슬롯
func RetrieveEmptyTrayFromTable(slot model.Slot) error {
	log.Infof("[PLC] 테이블 빈 트레이 회수. slot: %v", slot)

	if err := robot.JobRetrieveEmptyTrayFromTable(slot); err != nil {
		return err
	}

	return nil
}

// InputItem
//
// 물품 수납.
//
// - slot: 물품을 수납할 슬롯
func InputItem(slot model.Slot) error {
	log.Infof("[PLC] 물품 수납. 수납할 슬롯: %v", slot)

	if err := robot.JobInputItem(slot); err != nil {
		return err
	}

	return nil
}

// OutputItem
//
// 물품 불출.
// - slot: 물품을 꺼내올 슬롯
func OutputItem(slot model.Slot) error {
	log.Infof("[PLC] 물품 불출 시작. 꺼내올 슬롯 id=%v", slot.SlotId)

	if err := robot.JobOutputItem(slot); err != nil {
		return err
	}

	log.Infof("[PLC] 물품 불출 완료. 꺼내온 슬롯 id=%v", slot.SlotId)
	return nil
}
