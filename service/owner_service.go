package service

import (
	"APCS/config"
	"APCS/data/request"
	"APCS/data/response"
	"APCS/repository"
)

type OwnerService struct {
	OwnerRepository repository.OwnerRepository
}

func (o *OwnerService) InitService() error {
	db := config.DBConnection()
	o.OwnerRepository.AssignDB(db)

	return nil
}

func (o *OwnerService) CheckOwnerMatch(req request.OwnerReadRequest) (response.OwnerReadResponse, error) {
	resp, err := o.OwnerRepository.SelectOwnerByOwnerInfo(req)
	// if null 이면 재입력하라는 msg 보냄 - null이면 정보불일치 알림 전송
	return resp, err
}
