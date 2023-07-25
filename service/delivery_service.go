package service

import (
	"APCS/config"
	"APCS/data/response"
	"APCS/repository"
)

type DeliveryService struct {
	DliveryRepository repository.DeliveryRepository
}

func (d *DeliveryService) InitService() error {
	db := config.DBConnection()

	d.DliveryRepository.AssignDB(db)
	return nil
}

func (d *DeliveryService) CheckDeliveryMatch(deliveryId int) (*response.DeliveryReadResponse, error) {
	resp, err := d.DliveryRepository.SelectDeliveryByDeliveryId(deliveryId)
	// if null 이면 재입력하라는 msg 보냄 - null이면 정보불일치 알림 전송
	return resp, err
}
