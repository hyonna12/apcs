package request

type ItemCreateRequest struct {
	ItemName       string
	ItemHeight     int
	TrackingNumber int
	DeliveryId     int
	OwnerId        int
}
