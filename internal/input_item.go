package internal

import (
	"APCS/data/request"
	"APCS/module"
	"APCS/service"
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

/*
	 func NewStartStorage(slotService service.SlotService) *InputItem {
		return &InputItem{SlotService: slotService}
	}
*/

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
		_ = resp
		// 최적의 빈트레이 선정
		s.DeliveryBoxService.SetUpDoor("뒷문", "열림") // 뒷문 열림
		service.MoveTray()                         // 슬롯에서 테이블로
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

	// 7. 수납 가능한 슬롯 조회
	available, _ := s.SlotService.FindAvailableSlotList(h)
	// 8. 최적의 슬롯 선정
	lane, floor := s.SlotService.ChoiceBestSlot(available)
	_ = lane
	_ = floor
	/*
		트레이 유무 확인
		문설정(뒷문, 열림)
		이동(트레이, 테이블 -> 슬롯)
		문설정(뒷문, 닫힘)
		알림(수납완료)
	*/

}
