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
	item := request.ItemCreateRequest{ItemName: "6", TrackingNumber: 666, OwnerId: 1}
	delivery := request.DeliveryCreateRequest{DeliveryName: "1", PhoneNum: "01011111111", DeliveryCompany: "a"}

	n.InputItem.InputItem(delivery, item)
}

func main() {
	new := New{}
	new.Start()
}
