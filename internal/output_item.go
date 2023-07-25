package internal

import "APCS/service"

type OutputItem struct {
	DeliveryBoxService service.DeliveryBoxService
	RobotService       service.RobotService
	ItemServie         service.ItemService
	TrayServie         service.TrayService
	OwnerServie        service.OwnerService
	DeliveryService    service.DeliveryService
	SlotServie         service.SlotService
}

func (r *OutputItem) StartRelease() {}
