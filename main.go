package main

import (
	"APCS/data/request"
	"APCS/internal"
	log "github.com/sirupsen/logrus"
)

func init() {

	// logger 세팅
	// 파일명/호출 라인 표시
	log.SetReportCaller(true)
	// JSON 출력
	log.SetFormatter(&log.JSONFormatter{})

}

type New struct {
	InputItem  internal.InputItem
	OutputItem internal.OutputItem
}

func (n *New) Start() {
	log.Info("server start")

	/* item := request.ItemCreateRequest{ItemName: "2", TrackingNumber: 2, OwnerId: 1}
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
