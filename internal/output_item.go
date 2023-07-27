package internal

import (
	"APCS/plc"
	"APCS/service"
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

func (r *OutputItem) OutputItem() {}
