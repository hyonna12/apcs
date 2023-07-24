package model

import "time"

type Delivery struct {
	DeliveryId      int
	DeliveryName    string
	PhoneNum        string
	DeliveryCompany string
	CDatetime       time.Time
	UDatetime       time.Time
}
