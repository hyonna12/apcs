package plc

import "fmt"

type Robot struct{}

func (r *Robot) ScanTrayQrcode() {
	// 트레이 큐알코드 스캔하라는 요청
	// 파라미터 - lane, floor, robot_id
	// 반환값 - qrcode
}

func (r *Robot) MoveTray(from_lane, from_floor, to_lane, to_floor int) {
	// 트레이를 이동하라는 요청
	// 파라미터 - (lane, floor), (lane, floor), robot_id
	fmt.Println(from_lane, from_floor, "에서", to_lane, to_floor, "로 이동")
}

func (r *Robot) MoveRobot() {
	// 로봇을 이동하라는 요청
	// 파라미터 - (lane, floor), (lane, floor), robot_id
}
