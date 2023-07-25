package service

type RobotService struct{}

func ScanTrayQrcode() {
	// 트레이 큐알코드 스캔하라는 요청
	// 파라미터 - lane, floor, robot_id
	// 반환값 - qrcode
}

func MoveTray() {
	// 트레이를 이동하라는 요청
	// 파라미터 - (lane, floor), (lane, floor), robot_id
	// 트레이 db update
}

func MoveRobot() {
	// 로봇을 이동하라는 요청
	// 파라미터 - (lane, floor), (lane, floor), robot_id
}
