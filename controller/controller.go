package controller

import "APCS/service"

type Controller struct {
	DeliveryBoxService service.DeliveryBoxService
	RobotService       service.RobotService
	ItemServie         service.ItemService
	TrayServie         service.TrayService
	OwnerServie        service.OwnerService
	DeliveryService    service.DeliveryService
	SlotServie         service.SlotService
}

func Storage() {
}
func Release() {
}
func Sort() {
}
func OverallSort() {
}
