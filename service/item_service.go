package service

import (
	"APCS/config"
	"APCS/data/request"
	"APCS/data/response"
	"APCS/repository"
)

type ItemService struct {
	ItemRepository repository.ItemRepository
	SlotRepository repository.SlotRepository
}

func (i *ItemService) InitService() error {
	db := config.DBConnection()
	i.ItemRepository.AssignDB(db)

	return nil
}

func (i *ItemService) FindItemLocation(itemId int) (*response.SlotReadResponse, error) {
	resp, err := i.SlotRepository.SelectItemLocationByItemId(itemId)

	return resp, err
}

func (i *ItemService) CreateItemInfo(req request.ItemCreateRequest) error {
	_, err := i.ItemRepository.InsertItem(req)

	return err
}

func (i *ItemService) ChoiceItem() {
	// 알고리즘 돌려서 정리할때 가장 최적의 물품 선정
	// 파라미터 item 정보
}

func (i *ItemService) CheckItemMatch(lane, floor int) {
	i.ItemRepository.SelectItemBySlot(lane, floor)
	//아이템 입력받아서 db값과 일치하는지 확인
}
