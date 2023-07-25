package service

import (
	"APCS/config"
	"APCS/data/response"
	"APCS/repository"
)

type TrayService struct {
	TrayRepository *repository.TrayRepository
}

func (t *TrayService) InitService() error {
	db := config.DBConnection()

	t.TrayRepository = &repository.TrayRepository{}
	t.TrayRepository.AssignDB(db)

	return nil
}

func (t *TrayService) FindTrayList() (*[]response.TrayReadResponse, error) {
	resp, err := t.TrayRepository.SelectTrayList()
	return resp, err
}

func (t *TrayService) ChoiceEmptyTray() (*[]response.TrayReadResponse, error) {
	// 빈트레이 리스트
	resp, err := t.TrayRepository.SelectEmptyTrayList()
	// 알고리즘 돌려서 가장 최적의 빈트레이 선정하기
	return resp, err // 리턴값 바꿔야함 - 하나의 트레이

}
