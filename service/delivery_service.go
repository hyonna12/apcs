package service

import (
	"APCS/config"
	"APCS/data/request"
	"APCS/module"
	"APCS/repository"
	log "github.com/sirupsen/logrus"
)

type DeliveryService struct {
	DeliveryRepository repository.DeliveryRepository
	Notification       module.Notification
}

func (d *DeliveryService) InitService() error {
	db := config.DBConnection()

	d.DeliveryRepository.AssignDB(db)
	return nil
}

func (d *DeliveryService) CheckDeliveryMatch(req request.DeliveryCreateRequest) (bool, error) {
	resp, err := d.DeliveryRepository.SelectDeliveryByDeliveryInfo(req)
	log.Info(resp)
	// if null 이면 재입력하라는 msg 보냄 - null이면 정보불일치 알림 전송
	a := true
	if a {
		// 일치하면 다음 동작
	} else {
		// 알림전송 후 종료? 재입력?
		d.Notification.PushNotification("재입력")
	}
	return a, err
}
