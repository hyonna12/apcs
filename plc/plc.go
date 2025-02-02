package plc

import (
	"apcs_refactored/config"
	"apcs_refactored/messenger"
	"apcs_refactored/model"
	"apcs_refactored/plc/conn"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/robot"
	"apcs_refactored/plc/sensor"
	"apcs_refactored/plc/trayBuffer"
	"database/sql"
	"time"

	"apcs_refactored/interfaces"

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

// TroubleHandler PLC 트러블 이벤트 처리기
type TroubleHandler struct {
	kiosk interfaces.KioskView
}

// HandleTroubleEvent implements trouble.TroubleEventHandler
func (h *TroubleHandler) HandleTroubleEvent(troubleType string, details map[string]interface{}) error {
	log.Infof("[PLC] 트러블 발생: %s", troubleType)

	switch troubleType {
	case "trouble_event":
		if troubleKind, ok := details["trouble_type"].(string); ok {
			switch troubleKind {
			case "화재":
				log.Warn("[PLC] 화재 발생")
				if err := h.kiosk.ChangeView("/error/trouble"); err != nil {
					return err
				}
				return h.kiosk.SendEvent("화재")

			case "물품 끼임":
				log.Warn("[PLC] 물품 끼임 발생")
				if err := h.kiosk.ChangeView("/error/trouble"); err != nil {
					return err
				}
				return h.kiosk.SendEvent("물품 끼임")

			case "물품 낙하":
				log.Warn("[PLC] 물품 낙하 발생")
				if err := h.kiosk.ChangeView("/error/trouble"); err != nil {
					return err
				}
				return h.kiosk.SendEvent("물품 낙하")

			case "이물질 감지":
				log.Warn("[PLC] 이물질 감지")
				if err := h.kiosk.ChangeView("/error/trouble"); err != nil {
					return err
				}
				return h.kiosk.SendEvent("이물질 감지")
			}
		}

	case "network_status":
		if status, ok := details["status"].(string); ok && status == "disconnected" {
			log.Warn("[PLC] 네트워크 절체 감지")
			if err := h.kiosk.ChangeView("/error/trouble"); err != nil {
				return err
			}
			return h.kiosk.SendEvent("네트워크 절체")
		}

	default:
		log.Warnf("[PLC] 알 수 없는 트러블 타입: %s", troubleType)
	}

	return nil
}

func StartPlcClient(n *messenger.Node, kiosk interfaces.KioskView) {
	msgNode = n

	// 시뮬레이터 딜레이 설정
	simulatorDelay = time.Duration(config.Config.Plc.Simulation.Delay)

	// PLC 서버 연결 시작 및 클라이언트 가져오기
	plcClient := conn.ConnectPlcServer()
	if plcClient == nil {
		log.Error("[PLC] Failed to initialize PLC client")
		return
	}

	// 트러블 핸들러 설정
	handler := &TroubleHandler{kiosk: kiosk}
	plcClient.SetTroubleHandler(handler)

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
	commandId := robot.GenerateCommandId() // commandId 생성
	err := door.SetUpDoor(doorType, doorOperation, commandId)
	if err != nil {
		return err
	}
	// 완료 대기
	if err := robot.CheckCompletePlc(commandId); err != nil {
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
	commandId := robot.GenerateCommandId() // commandId 생성
	err := trayBuffer.SetUpTrayBuffer(bufferOperation, commandId)
	if err != nil {
		log.Error("버퍼 조작 에러", err)
		return err
	}
	// 완료 대기
	if err := robot.CheckCompletePlc(commandId); err != nil {
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

	trayId := trayBuffer.Buffer.Peek().(int64)
	//TrayIdOnTable = sql.NullInt64{Valid: false} // set null

	return trayId, nil
}

// InputItem
//
// 물품 수납. 트레이 id 반환
//
// - slot: 물품을 수납할 슬롯
func InputItem(slot model.Slot) error {
	log.Infof("[PLC] 물품 수납. slotId=%v", slot.SlotId)

	if err := robot.JobInputItem(slot); err != nil {
		return err
	}

	//trayIdOnTable := sql.NullInt64{Valid: false} // set null

	return nil
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

// ReturnItem
//
// 물품 재수납.
//
// - slot: 물품을 꺼내올 슬롯	***************
func ReturnItem(slot model.Slot, robotId int) error {
	log.Infof("[PLC] 물품 재수납 시작. 수납할 슬롯 id=%v", slot.SlotId)

	if err := robot.JobReturnItem(slot, robotId); err != nil {
		return err
	}

	log.Infof("[PLC] 물품 재수납 완료. 슬롯 id=%v", slot.SlotId)
	return nil
}
