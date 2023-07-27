package main

import (
	"APCS/data/request"
	"APCS/internal"
	"fmt"
)

type New struct {
	InputItem  internal.InputItem
	OutputItem internal.OutputItem
}

func (n *New) Start() {
	fmt.Println("server start")
	/* item := request.ItemCreateRequest{ItemName: "6", TrackingNumber: 666, OwnerId: 1}
	// owner 정보와 item 정보 분리하기!!
	delivery := request.DeliveryCreateRequest{DeliveryName: "1", PhoneNum: "01011111111", DeliveryCompany: "a"}

	n.InputItem.InputItem(delivery, item) */

	owner := request.OwnerReadRequest{OwnerName: "1", PhoneNum: "0109999", Address: "101-101"}
	n.OutputItem.OutputItem(owner)
}

func main() {
	new := New{}
	new.Start()
}
