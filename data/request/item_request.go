package request

import "time"

type ItemCreateRequest struct {
	ItemName       string
	ItemHeight     int
	TrackingNumber int
	InputDate      time.Time
	DeliveryId     int
	OwnerId        int
}
