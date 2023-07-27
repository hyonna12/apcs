package internal

import (
	"APCS/data/request"
	"APCS/plc"
	"APCS/service"
	"fmt"
	"sort"
)

type OutputItem struct {
	SensorPlc       plc.SensorPlc
	GatePlc         plc.GatePlc
	RobotPlc        plc.RobotPlc
	ItemServie      service.ItemService
	TrayServie      service.TrayService
	OwnerServie     service.OwnerService
	DeliveryService service.DeliveryService
	SlotServie      service.SlotService
}

func (o *OutputItem) OutputItem(owner request.OwnerReadRequest) {
	o.ItemServie.InitService()
	o.TrayServie.InitService()
	o.OwnerServie.InitService()
	o.DeliveryService.InitService()
	o.SlotServie.InitService()

	// 1. 해당 유저의 물품 정보 조회
	ownerInfo, _ := o.OwnerServie.CheckOwnerMatch(owner)
	ItemInfo, _ := o.ItemServie.ItemRepository.SelectItemListByOwnerId(ownerInfo.OwnerId)
	fmt.Println(ItemInfo)
	// 여러개인 경우 선택할 수 있도록??

	// 2. 테이블에 빈 트레이 유무 감지
	tableTray := o.SensorPlc.SenseTableForEmptyTray()
	fmt.Println("테이블에 빈 트레이 유무:", tableTray)
	if tableTray {
		trayId := 100 // db에 있는 트레이로!!
		// 트레이를 옮길 최적의 슬롯 찾기
		resp, _ := o.SlotServie.SlotRepository.SelectEmptySlotList()
		sort.SliceStable(resp, func(i, j int) bool {
			return resp[i].TransportDistance < resp[j].TransportDistance
		})
		fmt.Println(resp)
		trayLane := resp[0].Lane
		trayFloor := resp[0].Floor
		fmt.Println("트레이 옮길 슬롯:", trayLane, trayFloor)
		fmt.Println("빈트레이 옮기기")
		o.RobotPlc.MoveTray(0, 0, trayLane, trayFloor)
		// 슬롯 트레이 정보
		o.SlotServie.ChangeTrayInfo(trayLane, trayFloor, trayId)
	}
	// 3. 물품이 든 트레이 이동 / 4,5,6 하나로 묶을지?
	// 4. 뒷문 열림
	// 5. 물품 감지
	// 6. 뒷문 닫힘
	// 7. 앞문 열림
	// 8. 물품 감지
	// 9. 앞문 닫힘
	// 10. 트레이 테이블, 슬롯 테이블, 물품 테이블 업데이트
	// 11. 불출 완료 알림
}
