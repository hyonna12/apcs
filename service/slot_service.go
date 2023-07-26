package service

import (
	"APCS/config"
	"APCS/data/request"
	"APCS/data/response"
	"APCS/repository"
	"sort"
)

type SlotService struct {
	SlotRepository repository.SlotRepository
}

func (s *SlotService) InitService() error {
	db := config.DBConnection()
	s.SlotRepository.AssignDB(db)

	return nil
}

func (s *SlotService) FindSlotListForEmptyTray() (*[]response.SlotReadResponse, error) {
	resp, err := s.SlotRepository.SelectSlotListForEmptyTray()
	return resp, err
}

func (s *SlotService) FindSlotList() (*[]response.SlotReadResponse, error) {
	resp, err := s.SlotRepository.SelectSlotList()
	return resp, err
}

func (s *SlotService) FindAvailableSlotList(itemHeight int) (*[]response.SlotReadResponse, error) {
	resp, err := s.SlotRepository.SelectAvailableSlotList(itemHeight)
	return resp, err
}

func (s *SlotService) ChoiceSlot() {
	// 알고리즘 돌려서 정리 시 이동할 슬롯 찾기
}

func (s *SlotService) ChoiceOverallSortSlot() {
	// 알고리즘 돌려서 통합정리 시 이동할 슬롯 찾기
}

func (s *SlotService) ChoiceBestSlot(availableSlots *[]response.SlotReadResponse) (lane, floor int) {
	// available slots 받아서 알고리즘으로 수납할 최적의 슬롯 찾기
	list := *availableSlots
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].TransportDistance > list[j].TransportDistance
	})
	lane = list[0].Lane
	floor = list[0].Floor
	return lane, floor
}

func (s *SlotService) ChangeItemInfo(req request.SlotUpdateRequest) {
	s.SlotRepository.UpdateSlot(req)
}
