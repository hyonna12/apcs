package robot

import (
	"apcs_refactored/config"
	"apcs_refactored/model"
	"time"

	log "github.com/sirupsen/logrus"
)

type robotStatus string

const (
	robotStatusWaiting   robotStatus = "robotStatusWaiting"
	robotStatusAvailable robotStatus = "robotStatusAvailable"
	robotStatusWorking   robotStatus = "robotStatusWorking"
	robotStatusSlot      robotStatus = "robotStatusSlot"
)

type robot struct {
	id     int
	status robotStatus
	// job - 현재 수행하고 있는 job
	job          *job
	homePosition Position // 대기 위치 추가
}

type Position struct {
	x int
	z int
}

var (
	robots         []*robot
	simulatorDelay time.Duration
)

func InitRobots() {
	robotNumber := config.Config.Plc.Resource.Robot.Number
	robots = make([]*robot, robotNumber)
	// 시뮬레이터 딜레이 설정
	simulatorDelay = time.Duration(config.Config.Plc.Simulation.Delay)

	// TODO - 각 로봇과 통신 후 로봇 인스턴스 생성 및 등록
	log.Infof("[PLC_로봇] 로봇 통신 테스트 및 초기화")
	for i := range robots {
		robotConfig := config.Config.Plc.Resource.Robot.Robots[i]
		robots[i] = &robot{
			id:     robotConfig.ID,
			status: robotStatusAvailable,
			homePosition: Position{
				x: robotConfig.Home.X,
				z: robotConfig.Home.Z,
			},
		}
	}
}

func (r *robot) changeStatus(robotStatus robotStatus) {
	r.status = robotStatus
	go DistributeJob()
}

func (r *robot) moveToSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯으로 이동. robotId=%v, slotId=%v", r.id, slot.SlotId)

	// TODO - 슬롯으로 이동

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) moveToTable() error {
	log.Infof("[PLC_로봇_Step] 테이블로 이동. robotId=%v", r.id)
	// TODO - 테이블로 이동

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) pullFromSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯에서 트레이 꺼내기. robotId=%v, slotId=%v", r.id, slot.SlotId)
	// TODO - 슬롯에서 트레이 꺼내기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 500 * time.Millisecond)
	return nil
}

func (r *robot) pushToSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯으로 트레이 넣기. robotId=%v, slotId=%v", r.id, slot.SlotId)
	// TODO - 슬롯으로 트레이 넣기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) pullFromTable() error {
	log.Infof("[PLC_로봇_Step] 테이블에서 트레이 꺼내기. robotId=%v", r.id)
	// TODO - 테이블에서 트레이 꺼내기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) pushToTable() error {
	log.Infof("[PLC_로봇_Step] 테이블에 트레이 올리기. robotId=%v", r.id)
	// TODO - 테이블에 트레이 올리기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) scanTray() (int, error) {
	log.Infof("[PLC_로봇_Step] 트레이 QR코드 스캔. robotId=%v", r.id)
	// PLC 로봇 트레이 스캔
	TrayId := 20
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return TrayId, nil
}

// 대기 위치로 복귀
func (r *robot) returnToHome() error {
	log.Infof("[PLC_로봇_Step] 대기 위치로 복귀. robotId=%v, target=(x:%v,z:%v)", r.id, r.homePosition.x, r.homePosition.z)

	// 대기 위치로 이동
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}
