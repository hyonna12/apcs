package robot

import (
	"apcs_refactored/config"
	"apcs_refactored/model"
	log "github.com/sirupsen/logrus"
	"time"
)

type robotStatus string

const (
	robotStatusWaiting   robotStatus = "robotStatusWaiting"
	robotStatusAvailable robotStatus = "robotStatusAvailable"
	robotStatusWorking   robotStatus = "robotStatusWorking"
)

type robot struct {
	id     int
	status robotStatus
	// job - 현재 수행하고 있는 job
	job *job
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
	for i, _ := range robots {
		robots[i] = &robot{
			id:     i, // TODO - temp robot id
			status: robotStatusAvailable,
		}
	}
}

func (r *robot) changeStatus(robotStatus robotStatus) {
	r.status = robotStatus
	go DistributeJob()
}

func (r *robot) moveToSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯으로 이동. robot: %v, slot: %v", r, slot)

	// TODO - 슬롯으로 이동

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 1 * time.Second)

	return nil
}

func (r *robot) moveToTable() error {
	log.Infof("[PLC_로봇_Step] 테이블로 이동. robot: %v", *r)
	// TODO - 테이블로 이동

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 1 * time.Second)

	return nil
}

func (r *robot) pullFromSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯에서 트레이 꺼내기. robot: %v, slot: %v", *r, slot)
	// TODO - 슬롯에서 트레이 꺼내기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 1 * time.Second)
	return nil
}

func (r *robot) pushToSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯으로 트레이 넣기. robot: %v, slot: %v", *r, slot)
	// TODO - 슬롯으로 트레이 넣기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 1 * time.Second)

	return nil
}

func (r *robot) pullFromTable() error {
	log.Infof("[PLC_로봇_Step] 테이블에서 트레이 꺼내기. robot: %v", *r)
	// TODO - 테이블에서 트레이 꺼내기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 1 * time.Second)

	return nil
}

func (r *robot) pushToTable() error {
	log.Infof("[PLC_로봇_Step] 테이블에 트레이 올리기. robot: %v", *r)
	// TODO - 테이블에 트레이 올리기

	// TODO - temp - 시뮬레이터
	time.Sleep(simulatorDelay * 1 * time.Second)

	return nil
}
