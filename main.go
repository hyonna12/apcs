package main

import (
	"APCS/data/request"
	"APCS/internal"
	"fmt"
)

type New struct {
	InputItem internal.InputItem
}

func (n *New) Start() {
	fmt.Println("server start")
	item := request.ItemCreateRequest{ItemName: "10", TrackingNumber: 1111, OwnerId: 10}
	delivery := request.DeliveryCreateRequest{DeliveryName: "10", PhoneNum: "010", DeliveryCompany: "10"}

	n.InputItem.StartStorage(delivery, item)
}

func main() {
	new := New{}
	new.Start()
}
