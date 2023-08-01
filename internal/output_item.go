package internal

import (
	"APCS/data/request"
	"APCS/data/response"
	"APCS/plc"
	"APCS/service"
	log "github.com/sirupsen/logrus"
	"sort"
)

type OutputItem struct {
	SensorPlc       plc.SensorPlc
	GatePlc         plc.GatePlc
	RobotPlc        plc.RobotPlc
	ItemService     service.ItemService
	TrayService     service.TrayService
	OwnerService    service.OwnerService
	DeliveryService service.DeliveryService
	SlotService     service.SlotService
}

func (o *OutputItem) OutputItem(owner request.OwnerReadRequest) {
	o.ItemService.InitService()
	o.TrayService.InitService()
	o.OwnerService.InitService()
	o.DeliveryService.InitService()
	o.SlotService.InitService()

	// 1. 해당 유저의 물품 정보 조회
	ownerInfo, _ := o.OwnerService.CheckOwnerMatch(owner)
	log.Info("유저 조회:", ownerInfo)
	ItemInfo, _ := o.ItemService.ItemRepository.SelectItemListByOwnerId(ownerInfo.OwnerId)
	log.Info("해당 유저의 아이템 리스트:", ItemInfo)

	// 불출할 물품 선택(한 개인 경우/여러 개 중 하나만 꺼냄/여러 개 중 여러 개 꺼냄)
	// 키오스크 - 물품 선택
	OutputItem := response.ItemReadResponse{ItemId: 2, ItemName: "2", ItemHeight: 3, Lane: 2, Floor: 3, TrayId: 7}
	log.Info(OutputItem)

	// 2. 테이블에 빈 트레이 유무 감지
	tableTray := o.SensorPlc.SenseTableForEmptyTray()
	log.Info("테이블에 빈 트레이 유무:", tableTray)
	if tableTray {
		trayId := 11 // db에 있는 트레이로!!
		// 트레이를 옮길 최적의 슬롯 찾기
		resp, _ := o.SlotService.SlotRepository.SelectEmptySlotList()
		sort.SliceStable(resp, func(i, j int) bool {
			return resp[i].TransportDistance < resp[j].TransportDistance
		})
		log.Info(resp)
		trayLane := resp[0].Lane
		trayFloor := resp[0].Floor
		log.Info("트레이 옮길 슬롯:", trayLane, trayFloor)
		log.Info("빈트레이 옮기기")
		o.RobotPlc.MoveTray(0, 0, trayLane, trayFloor)
		// 슬롯 트레이 정보
		o.SlotService.ChangeTrayInfo(trayLane, trayFloor, trayId)
	}

	// 3. 뒷문 열림
	o.GatePlc.SetUpDoor("뒷문", "열림")
	// 4. 물품이 든 트레이 이동 / 4,5,6 하나로 묶을지?
	o.RobotPlc.MoveTray(OutputItem.Lane, OutputItem.Floor, 0, 0)
	// 5. 물품 감지
	result := o.SensorPlc.SenseTableForItem()
	if !result {
		log.Info("물품 들어올 때까지 대기")
	}
	// 6. 뒷문 닫힘
	o.GatePlc.SetUpDoor("뒷문", "닫힘")
	// 7. 앞문 열림
	o.GatePlc.SetUpDoor("앞문", "열림")
	// 8. 물품 감지
	result = o.SensorPlc.SenseTableForItem()
	if result {
		log.Info("물품 가져갈 때까지 대기")
	}
	// 9. 앞문 닫힘
	o.GatePlc.SetUpDoor("앞문", "닫힘")

	// 10. 트레이 테이블, 슬롯 테이블, 물품 테이블 업데이트, 같은 행 keep_cnt
	tray_info := request.TrayUpdateRequest{ItemId: 0, TrayOccupied: true}
	o.TrayService.UpdateTray(OutputItem.TrayId, tray_info)

	o.ItemService.ItemRepository.UpdateOutputTime(OutputItem.ItemId)
	o.SlotService.ChangeTrayInfo(OutputItem.Lane, OutputItem.Floor, 0)

	Req := request.SlotUpdateRequest{SlotEnabled: false, SlotKeepCnt: 0, Lane: OutputItem.Lane, Floor: OutputItem.Floor}
	o.SlotService.ChangeOutputSlotInfo(OutputItem.ItemHeight, Req)
	o.SlotService.SlotRepository.UpdateOutputSlotKeepCnt(OutputItem.Lane, OutputItem.Floor)
	// 11. 불출 완료 알림
	o.DeliveryService.Notification.PushNotification("불출 완료")
}
