package service

import (
	"APCS/config"
	"APCS/data/request"
	"APCS/data/response"
	"APCS/repository"
	"fmt"
	"sort"

	"golang.org/x/exp/slices"
)

type SlotService struct {
	SlotRepository repository.SlotRepository
}

func (s *SlotService) InitService() error {
	db := config.DBConnection()
	s.SlotRepository.AssignDB(db)

	return nil
}

func (s *SlotService) FindSlotListForEmptyTray() ([]response.SlotReadResponse, error) {
	resp, err := s.SlotRepository.SelectSlotListForEmptyTray()

	return resp, err
}

func (s *SlotService) FindSlotList() (*[]response.SlotReadResponse, error) {
	resp, err := s.SlotRepository.SelectSlotList()
	return resp, err
}

func (s *SlotService) FindAvailableSlotList(itemHeight int) ([]response.SlotReadResponse, error) {
	resp, err := s.SlotRepository.SelectAvailableSlotList(itemHeight)
	return resp, err
}

func (s *SlotService) ChoiceSlot() {
	// 알고리즘 돌려서 정리 시 이동할 슬롯 찾기
}

func (s *SlotService) ChoiceOverallSortSlot() {
	// 알고리즘 돌려서 통합정리 시 이동할 슬롯 찾기
}

func (s *SlotService) ChoiceBestSlot(availableSlots []response.SlotReadResponse) (lane, floor int) {
	// available slots 받아서 알고리즘으로 수납할 최적의 슬롯 찾기
	sort.SliceStable(availableSlots, func(i, j int) bool {
		return availableSlots[i].TransportDistance < availableSlots[j].TransportDistance
	})
	lane = availableSlots[0].Lane
	floor = availableSlots[0].Floor
	return lane, floor
}

func (s *SlotService) FindStorageSlotWithTray(itemHeight, lane, floor int) ([]response.SlotReadResponse, error) {
	resp, err := s.SlotRepository.SelectStorageSlotListWithTray(itemHeight, lane, floor)
	return resp, err
}
func (s *SlotService) FindEmptySlotList(best_lane, best_floor, item_height int) ([]response.SlotReadResponse, error) {
	var list []response.SlotReadResponse
	Resp, _ := s.SlotRepository.SelectEmptySlotList() // 트레이가능슬롯들
	// 수납슬롯들
	var slots []response.SlotReadResponse
	for z := 0; z < item_height; z++ {
		slot, _ := s.SlotRepository.SelectSlotInfoByLocation(best_lane, best_floor-z)
		slots = append(slots, slot)
	}
	fmt.Println("수납슬롯들:", slots)

	for _, i := range Resp {
		if !slices.Contains(slots, i) {
			list = append(list, i)
		}
	}
	fmt.Println("트레이슬롯들", list)

	// empty 슬롯 리스트 뽑아서 available이 아닌거 for 로 담기
	return list, nil
}

func (s *SlotService) ChangeSlotInfo(lane, floor, tray_id, item_id int) {
	update := request.SlotUpdateRequest{SlotEnabled: false, SlotKeepCnt: 0, TrayId: tray_id, ItemId: item_id, Lane: lane, Floor: floor}
	// 슬롯 정보 가져와서
	s.SlotRepository.UpdateSlot(update)
}

func (s *SlotService) ChangeStorageSlotInfo(itemHeight, lane, floor, item_id int) {
	resp, _ := s.SlotRepository.SelectSlotInfoByLocation(lane, floor)
	fmt.Println(resp)
	update := request.SlotUpdateRequest{SlotEnabled: false, SlotKeepCnt: 2, ItemId: item_id, Lane: lane, Floor: floor}
	fmt.Println(update)
	// 슬롯 정보 가져와서
	s.SlotRepository.UpdateStorageSlotList(itemHeight, update)
}

func (s *SlotService) ChangeOutputSlotInfo(itemHeight int, req request.SlotUpdateRequest) {
	update := request.SlotUpdateRequest{SlotEnabled: true, Lane: req.Lane, Floor: req.Floor}
	fmt.Println(update)
	// 슬롯 정보 가져와서
	s.SlotRepository.UpdateOutputSlotList(itemHeight, update)
	s.SlotRepository.UpdateOutputSlotListKeepCnt(itemHeight, req.Lane, req.Floor)
}

func (s *SlotService) ChangeTrayInfo(lane, floor, tray_id int) {
	// 슬롯 정보 가져와서
	if tray_id == 0 {
		fmt.Println(lane, floor, tray_id)
		s.SlotRepository.UpdateSlotToEmptyTray(lane, floor)
	} else {
		s.SlotRepository.UpdateSlotTrayInfo(lane, floor, tray_id)
	}
}

func (s *SlotService) ChangeItemInfo(lane, floor, item_id int) {
	// 슬롯 정보 가져와서
	s.SlotRepository.UpdateSlotItemInfo(lane, floor, item_id)
}
