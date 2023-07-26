package internal

import (
	"APCS/data/request"
	"APCS/service"
)

type InputItem struct {
	DeliveryBoxService service.DeliveryBoxService
	RobotService       service.RobotService
	ItemService        service.ItemService
	TrayService        service.TrayService
	OwnerService       service.OwnerService
	DeliveryService    service.DeliveryService
	SlotService        service.SlotService
}

/*
	 func NewStartStorage(slotService service.SlotService) *InputItem {
		return &InputItem{SlotService: slotService}
	}
*/

func (s *InputItem) StartStorage(req request.DeliveryCreateRequest) {
	s.ItemService.InitService()
	s.TrayService.InitService()
	s.DeliveryService.InitService()
	s.SlotService.InitService()
	s.OwnerService.InitService()

	// 물품기사 일치여부 확인
	s.DeliveryService.CheckDeliveryMatch(req)

	// 테이블에 빈 트레이 유무 감지
	s.DeliveryBoxService.SenseTableForEmptyTray()
	b := true
	if b {
		// 있으면 다음 동작
	} else {
		resp, _ := s.SlotService.FindSlotListForEmptyTray()
		_ = resp
		// 최적의 빈트레이 선정
		s.DeliveryBoxService.SetUpDoor("뒷문", "열림") // 뒷문 열림
		service.MoveTray()                         // 슬롯에서 테이블로
		s.DeliveryBoxService.SetUpDoor("뒷문", "닫힘") // 뒷문 닫힘

	}

	/* 빈트레이가 있는 슬롯리스트 조회
				빈 트레이 있음
					빈 트레이 선정
					문설정(뒷문, 열림)
					이동(트레이, 슬롯 -> 테이블)
					문설정(뒷문, 닫힘)
				빈 트레이 없음(보류)
					문설정(앞문, 열림)
					(물품 기사가 셀프로 트레일 수납)보류
	문설정(앞문, 열림)
	물품 정보 입력(바코드, 수기 등)
		물품 정보 생성
	수납 완료 버튼
	문설정(앞문, 닫힘)
	물품 정보(높이, 무게, 돌출) 감지
		수납 불가
			알림(문제점)
			문설정(앞문, 열림)
			재수납 버튼 클릭
				문설정(앞문, 닫힘)
				물품 정보(높이, 무게, 돌출) 감지
			수납 취소 버튼 클릭
				수납 중단
	수납 가능한 슬롯 조회
	최적의 슬롯 선정
	문설정(뒷문, 열림)
	이동(트레이, 테이블 -> 슬롯)
	문설정(뒷문, 닫힘)
	알림(수납완료) */

}
