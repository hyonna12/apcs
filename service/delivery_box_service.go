package service

import "fmt"

type DeliveryBoxService struct {
	SlotService
}

func (d *DeliveryBoxService) SetUpDoor(gate, operate string) {
	// 문을 제어하는 요청
	// 파라미터 - frontgate/backgate, open/close
	fmt.Println(gate, operate)

}

func (d *DeliveryBoxService) SenseTableForTrayWithItem() {
	// 테이블에 물품있는 트레이 유무
}

func (d *DeliveryBoxService) SenseTableForItem() {
	// 테이블에 물품 유무
}

func (d *DeliveryBoxService) SenseTableForEmptyTray() {
	// 테이블에 빈 트레이 유무 감지
	// 센서 동작
	// 반환값 - true/false

}
func (d *DeliveryBoxService) SenseItemInfo() {
	// 물품정보(높이, 무게) 감지
	// 반환값 - 높이, 무게
}

func (d *DeliveryBoxService) SenseNotMoveForxm() {
	// 미이동 감지
}
