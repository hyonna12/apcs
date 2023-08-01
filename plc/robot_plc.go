package plc

import log "github.com/sirupsen/logrus"

type RobotPlc struct{}

func (r *RobotPlc) GetTraylocation() {
	// 트레이 큐알코드 스캔하라는 요청
	// 파라미터 - lane, floor, robot_id
	// 반환값 - qrcode
}

func (r *RobotPlc) MoveTray(from_lane, from_floor, to_lane, to_floor int) {
	// 트레이를 이동하라는 요청
	// 파라미터 - (lane, floor), (lane, floor), robot_id
	log.Info(from_lane, from_floor, "에서", to_lane, to_floor, "로 이동")
}

func (r *RobotPlc) MoveRobot(from_lane, from_floor, to_lane, to_floor int) {
	// 로봇을 이동하라는 요청
	// 파라미터 - (lane, floor), (lane, floor), robot_id
	log.Info(from_lane, from_floor, "에서", to_lane, to_floor, "로 이동")
}

func (r *RobotPlc) GetRobotState() {
	// 로봇 상태 조회 명령
}
