package service

import "fmt"

type DeliveryBoxService struct{}

// 문을 제어하는 요청
func (d *DeliveryBoxService) SetUpDoor(gate, operate string) {
	// 파라미터 - frontgate/backgate, open/close
	fmt.Println(gate, operate)

}

// 테이블에 물품있는 트레이 유무
func (d *DeliveryBoxService) SenseTableForTrayWithItem() bool {
	a := true
	return a
}

// 테이블에 물품 유무
func (d *DeliveryBoxService) SenseTableForItem() bool {
	a := true
	return a
}

// 테이블에 빈 트레이 유무 감지
func (d *DeliveryBoxService) SenseTableForEmptyTray() bool {
	// 센서 동작
	a := true
	return a

}

// 물품정보(높이, 무게) 감지
func (d *DeliveryBoxService) SenseItemInfo() (item_height, item_weight int) {
	// 센서 동작
	item_height = 3
	item_weight = 5

	return item_height, item_weight
}

// 미이동 감지
func (d *DeliveryBoxService) SenseNotMoveForxm() {
}
