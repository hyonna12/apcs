package robot

import (
	"apcs_refactored/config"
	"apcs_refactored/model"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
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

type RobotRequest struct {
	RobotId  int      `json:"robotId"`
	Location Location `json:"location"`
}
type Location struct {
	Lane  int `json:"x"`
	Floor int `json:"y"`
}
type RobotState struct {
	Id       int      `json:"id"`
	Location Location `json:"location"`
	IsTray   bool     `json:"isTray"`
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
			id:     i + 1, // TODO - temp robot id
			status: robotStatusAvailable,
		}
	}
}

func (r *robot) changeStatus(robotStatus robotStatus) {
	r.status = robotStatus
	go DistributeJob()
}

func (r *robot) moveToSlot(slot model.Slot) error {
	log.Infof("[PLC_로봇_Step] 슬롯으로 이동. robotId=%v, slotId=%v", r.id, slot.SlotId)

	// PLC 로봇 슬롯으로 이동
	data := RobotRequest{RobotId: r.id, Location: Location{slot.Lane, slot.Floor}}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	_, err := http.Post("http://localhost:8000/move/slot", "application/json", buff)
	if err != nil {
		return err
	}

	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) moveToTable() error {
	log.Infof("[PLC_로봇_Step] 테이블로 이동. robotId=%v", r.id)
	// PLC 로봇 테이블로 이동
	data := RobotRequest{RobotId: r.id}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	_, err := http.Post("http://localhost:8000/move/table", "application/json", buff)
	if err != nil {
		return err
	}

	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) pullTray() error {
	log.Infof("[PLC_로봇_Step] 트레이 꺼내기. robotId=%v", r.id)
	// PLC 로봇 트레이 꺼내기
	data := RobotRequest{RobotId: r.id}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	_, err := http.Post("http://localhost:8000/pull/tray", "application/json", buff)
	if err != nil {
		return err
	}

	time.Sleep(simulatorDelay * 500 * time.Millisecond)
	return nil
}

func (r *robot) pushTray() error {
	log.Infof("[PLC_로봇_Step] 트레이 넣기. robotId=%v", r.id)
	// PLC 로봇 트레이 넣기
	data := RobotRequest{RobotId: r.id}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	_, err := http.Post("http://localhost:8000/push/tray", "application/json", buff)
	if err != nil {
		return err
	}
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func (r *robot) scanTray() error {
	log.Infof("[PLC_로봇_Step] 트레이 QR코드 스캔. robotId=%v", r.id)
	// PLC 로봇 트레이 스캔
	data := RobotRequest{RobotId: r.id}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	_, err := http.Post("http://localhost:8000/scan/tray", "application/json", buff)
	if err != nil {
		return err
	}
	time.Sleep(simulatorDelay * 500 * time.Millisecond)

	return nil
}

func GetRobotState() ([]RobotState, error) {
	log.Infof("[PLC_로봇] 로봇 상태 조회")
	// PLC 로봇 상태 조회
	var robotState []RobotState
	resp, err := http.Get("http://localhost:8000/robot")
	if err != nil {
		return robotState, err
	}

	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return robotState, err
	}
	json.Unmarshal(respData, &robotState)

	log.Infof("[PLC_Robot] 로봇 상태: %v", robotState)

	return robotState, err
}
