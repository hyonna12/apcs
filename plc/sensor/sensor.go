package sensor

import (
	"math/rand"

	log "github.com/sirupsen/logrus"
)

var (
	// TODO - temp - 키오스크 임시 물건 꺼내기 버튼 시뮬레이션 용
	IsItemOnTable = false
)

type ItemDimension struct {
	Height      int
	Width       int
	Weight      int
	TrackingNum int
}

// SenseTableForItem
//
// 테이블에 물품이 있는지 감지.
// 있으면 true, 없으면 false 반환.
func SenseTableForItem() (bool, error) {
	log.Infof("[PLC_Sensor] 테이블 물품 존재 여부 감지: %v", IsItemOnTable)
	// TODO - PLC 센서 물품 존재 여부 감지
	// TODO - temp - 물건 꺼내기 버튼
	return IsItemOnTable, nil
}

// SenseItemInfo
//
// 테이블에 올려진 물품 크기/무게 계측.
func SenseItemInfo() (ItemDimension, error) {
	// TODO - temp
	itemDimension := ItemDimension{
		Height: rand.Intn(270) + 1,
		Width:  rand.Intn(10) + 1,
		Weight: rand.Intn(10) + 1,
	}
	log.Infof("[PLC_Sensor] 크기/무게 측정: %v", itemDimension)

	return itemDimension, nil
}
