package request

type DeliveryReadRequest struct {
	DeliveryId int
}

type DeliveryCreateRequest struct {
	DeliveryName    string
	PhoneNum        string
	DeliveryCompany string
}
