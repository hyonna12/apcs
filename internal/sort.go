package internal

import "APCS/service"

type Sort struct {
	DeliveryBoxService service.DeliveryBoxService
	RobotService       service.RobotService
	ItemServie         service.ItemService
	TrayServie         service.TrayService
	OwnerServie        service.OwnerService
	DeliveryService    service.DeliveryService
	SlotServie         service.SlotService
}

func (s *Sort) StartSort() {}