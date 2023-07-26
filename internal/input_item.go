package internal

import (
	"APCS/data/request"
	"APCS/module"
	"APCS/service"
	"sort"
)

type InputItem struct {
	DeliveryBoxService service.DeliveryBoxService
	RobotService       service.RobotService
	ItemService        service.ItemService
	TrayService        service.TrayService
	OwnerService       service.OwnerService
	DeliveryService    service.DeliveryService
	SlotService        service.SlotService
	Notification       module.Notification
}

func (s *InputItem) StartStorage(delivery request.DeliveryCreateRequest, item request.ItemCreateRequest) {
	s.ItemService.InitService()
	s.TrayService.InitService()
	s.DeliveryService.InitService()
	s.SlotService.InitService()
	s.OwnerService.InitService()

	// 1. 물품기사 일치여부 확인
	s.DeliveryService.CheckDeliveryMatch(delivery)

	// 2. 테이블에 빈 트레이 유무 감지
	result := s.DeliveryBoxService.SenseTableForEmptyTray()
	if result {
		// 있으면 다음 동작
	} else {
		resp, _ := s.SlotService.FindSlotListForEmptyTray()
		list := *resp
		// 슬롯 리스트 찾아서 정렬
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].TransportDistance > list[j].TransportDistance
		})

		tray_lane := list[0].Lane
		tray_floor := list[0].Floor
		// 최적의 빈트레이 선정
		s.DeliveryBoxService.SetUpDoor("뒷문", "열림")           // 뒷문 열림
		s.RobotService.MoveTray(tray_lane, tray_floor, 0, 0) // 슬롯에서 테이블로
		s.SlotService.ChangeTrayInfo(tray_lane, tray_floor, 0)
		s.DeliveryBoxService.SetUpDoor("뒷문", "닫힘") // 뒷문 닫힘
		// 슬롯에 빈트레이가 없으면 셀프 트레일 수납

	}

	// 3. 앞문 열림
	s.DeliveryBoxService.SetUpDoor("앞문", "열림")
	// 4. 물품 감지
	s.DeliveryBoxService.SenseTableForItem()
	// 5. 앞문 닫힘
	s.DeliveryBoxService.SetUpDoor("앞문", "닫힘")
	// 6. 물품 정보 감지
	h, w := s.DeliveryBoxService.SenseItemInfo()
	if w > 10 {
		s.DeliveryBoxService.SetUpDoor("앞문", "열림")
		s.Notification.PushNotification("무게 초과")
		// 수납 중단?
	}
	// 물품 정보 insert
	item.ItemHeight = h
	s.ItemService.CreateItemInfo(item)
	// 트레이 정보 update
	tray_id := 10
	item_id := 10
	tray := request.TrayUpdateRequest{TrayOccupied: true, ItemId: item_id}
	s.TrayService.UpdateTray(tray_id, tray)

	// 7. 수납 가능한 슬롯 조회
	available, _ := s.SlotService.FindAvailableSlotList(h)
	// 8. 최적의 슬롯 선정
	best_lane, best_floor := s.SlotService.ChoiceBestSlot(available)
	// 9. 최적 슬롯에 트레이 유무 확인
	resp, _ := s.SlotService.FindStorageSlotWithTray(h, best_lane, best_floor)
	best_slot := *resp
	if len(best_slot) != 0 {
		for _, num := range best_slot {
			// 트레이를 옮길 최적의 슬롯 찾기
			slots, _ := s.SlotService.FindEmptySlotList(best_lane, best_floor)
			list := *slots
			// 슬롯 리스트 찾아서 정렬
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].TransportDistance > list[j].TransportDistance
			})

			tray_lane := list[0].Lane
			tray_floor := list[0].Floor

			s.RobotService.MoveTray(num.Lane, num.Floor, tray_lane, tray_floor)
			// 슬롯 트레이 정보 update
			s.SlotService.ChangeTrayInfo(num.Lane, num.Floor, 0)
			s.SlotService.ChangeTrayInfo(tray_lane, tray_floor, num.TrayId)
		}

	}
	// 10. 뒷문 열림
	s.DeliveryBoxService.SetUpDoor("뒷문", "열림")
	// 11. 물품이 든 트레이 이동
	s.RobotService.MoveTray(0, 0, best_lane, best_floor)
	s.SlotService.ChangeTrayInfo(best_lane, best_floor, tray_id)
	s.SlotService.ChangeStorageSlotInfo(item.ItemHeight, best_lane, best_floor, item_id)
	s.SlotService.SlotRepository.UpdateStorageSlotKeepCnt(best_lane, best_floor)

	// 12. 뒷문 닫힘
	s.DeliveryBoxService.SetUpDoor("뒷문", "닫힘")
	// 13. 알림
	s.Notification.PushNotification("수납완료")

}
