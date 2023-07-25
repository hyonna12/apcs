package internal

import "APCS/service"

type InputItem struct {
	DeliveryBoxService service.DeliveryBoxService
	RobotService       service.RobotService
	ItemService        service.ItemService
	TrayService        service.TrayService
	OwnerService       service.OwnerService
	DeliveryService    service.DeliveryService
	SlotService        service.SlotService
}

func NewStartStorage(slotService service.SlotService) *InputItem {
	return &InputItem{SlotService: slotService}
}
func (s *InputItem) StartStorage(deliveryId int) {
	s.ItemService.InitService()
	s.TrayService.InitService()
	s.DeliveryService.InitService()
	s.SlotService.InitService()
	s.OwnerService.InitService()

	// 물품기사 일치여부 확인
	s.DeliveryService.CheckDeliveryMatch(deliveryId)
	// 테이블에 빈 트레이 유무 감지
	s.DeliveryBoxService.SenseTableForEmptyTray()

}
