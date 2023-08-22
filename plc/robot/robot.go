package robot

import (
	"apcs_refactored/config"
	"apcs_refactored/model"
	log "github.com/sirupsen/logrus"
)

type robotStatus string

const (
	waiting   robotStatus = "waiting"
	available robotStatus = "Available"
	working   robotStatus = "Working"
)

type robot struct {
	id     int
	status robotStatus
	// job - 현재 수행하고 있는 job
	job *job
}

var (
	robots []*robot
)

func InitRobots() {
	robotNumber := config.Config.Plc.Resource.Robot.Number
	robots = make([]*robot, robotNumber)

	// TODO - 각 로봇과 통신 후 로봇 인스턴스 생성 및 등록
	log.Infof("[PLC_로봇] 로봇 통신 테스트 및 초기화")
	for i, _ := range robots {
		robots[i] = &robot{
			id:     i, // TODO - temp robot id
			status: available,
		}
	}
}

func (r *robot) moveToSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯으로 이동. robot: %v, slot: %v", r, slot)
	// TODO - 슬롯으로 이동
	return nil
}

func (r *robot) moveToTable() error {
	log.Infof("[PLC_로봇_Step] 테이블로 이동. robot: %v", *r)
	// TODO - 테이블로 이동
	return nil
}

func (r *robot) pullFromSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯에서 트레이 꺼내기. robot: %v, slot: %v", *r, slot)
	// TODO - 슬롯에서 트레이 꺼내기
	return nil
}

func (r *robot) pushToSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯으로 트레이 넣기. robot: %v, slot: %v", *r, slot)
	// TODO - 슬롯으로 트레이 넣기
	return nil
}

func (r *robot) pullFromTable() error {
	log.Infof("[PLC_로봇_Step] 테이블에서 트레이 꺼내기. robot: %v", *r)
	// TODO - 테이블에서 트레이 꺼내기
	return nil
}

func (r *robot) pushToTable() error {
	log.Infof("[PLC_로봇_Step] 테이블에 트레이 올리기. robot: %v", *r)
	// TODO - 테이블에 트레이 올리기
	return nil
}
