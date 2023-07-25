package service

import (
	"APCS/config"
	"APCS/data/response"
	"APCS/repository"
)

type DeliveryService struct {
	Repository *repository.DeliveryRepository
}

func (d *DeliveryService) InitService() error {
	db := config.DBConnection()

	d.Repository = &repository.DeliveryRepository{}
	d.Repository.AssignDB(db)

	return nil
}

func (d *DeliveryService) CheckDeliveryMatch(deliveryId int) (*[]response.DeliveryReadResponse, error) {
	resp, err := d.Repository.SelectDeliveryByDeliveryId(deliveryId)
	// if null 이면 재입력하라는 msg 보냄
	return resp, err
}
