package response

import "time"

type DeliveryReadResponse struct {
	DeliveryId      int
	DeliveryName    string
	PhoneNum        string
	DeliveryCompany string
	CDatetime       time.Time
	UDatetime       time.Time
}
