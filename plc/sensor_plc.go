package plc

type SensorPlc struct{}

// 테이블 감지 한번에??
// 테이블에 빈 트레이 유무 감지
func (s *SensorPlc) SenseTableForEmptyTray() bool {
	// 센서 동작
	a := true
	return a
}

// 테이블에 물품있는 트레이 유무
func (s *SensorPlc) SenseTableForTrayWithItem() bool {
	a := true
	return a
}

// 테이블에 물품 유무
func (s *SensorPlc) SenseTableForItem() bool {
	a := true
	return a
}

// 물품정보(높이, 무게) 감지
func (s *SensorPlc) DetectBox() (item_height, item_weight int) {
	// 센서 동작
	item_height = 3
	item_weight = 5

	return item_height, item_weight
}

// 미이동 감지
func (s *SensorPlc) SenseNotMoveForxm() {
}
