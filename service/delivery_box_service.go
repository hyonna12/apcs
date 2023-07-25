package service

type DeliveryBoxService struct{}

func SetUpDoor() {
	// 문을 제어하는 요청
	// 파라미터 - frontgate/backgate, open/close
}

func SenseTableForTrayWithItem() {
	// 테이블에 물품있는 트레이 유무
}

func SenseTableForItem() {
	// 테이블에 물품 유무
}

func SenseTableForEmptyTray() {
	// 테이블에 빈 트레이 유무
}
func SenseItemInfo() {
	// 물품정보(높이, 무게) 감지
	// 반환값 - 높이, 무게
}
