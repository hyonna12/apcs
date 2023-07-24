package model

import "time"

type Item struct {
	ItemId         int
	ItemName       string
	ItemHeight     int
	TrackingNumber int
	InputDate      time.Time
	OutputDate     time.Time
	DeliveryId     int
	OwnerId        int
	CDatetime      time.Time
	UDatetime      time.Time
}
