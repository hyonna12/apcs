package robot

import (
	"apcs_refactored/model"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/resource"
	"apcs_refactored/plc/trayBuffer"
	"time"

	_ "github.com/future-architect/go-mcprotocol/mcp"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// job
//
// 한 로봇이 연속적으로 수행해야 하는 작업 단위.
//
// 예시: 빈 트레이 서빙(JobServeEmptyTrayToTable).
type job struct {
	// id - 식별용 uuid
	id string
	// robot job을 실행할 로봇
	robot *robot
	// requiredRobotStatus - job을 수행하기 위한 로봇의 상태 조건
	requiredRobotStatus robotStatus
	// robotWaiting - job에 로봇을 배정하는 채널
	robotWaiting chan *robot
	// timestamp - job이 생성된 시각
	timestamp   time.Time
	description string
}

var (
	jobQueue []*job
	RespPlc  interface{}
)

func newJob(robotStatus robotStatus, jobDescription string) *job {
	job := &job{
		id:                  uuid.NewString(),
		requiredRobotStatus: robotStatus,
		robotWaiting:        make(chan *robot),
		timestamp:           time.Now(),
		description:         jobDescription,
	}

	return job
}

// DistributeJob - 로봇에게 job을 배정함.
//
// 다음 두 경우에 실행됨.
//
// 1. 로봇의 상태가 변화했을 때(changeRobotStatus 함수)
//
// 2. 새로운 job이 추가되었을 때(getRobot 함수)
func DistributeJob() {
	if len(jobQueue) == 0 {
		return
	}

	// 로봇 상태를 구별하기 때문에 엄밀히는 queue가 아니지만, 상태에 따라 순서는 구분됨
	for i, job := range jobQueue {
		for _, robot := range robots {
			if robot.status == job.requiredRobotStatus && job.robot == nil {
				job.robot = robot
				job.robotWaiting <- robot
				// 로봇이 배정된 Job 삭제
				jobQueue = append(jobQueue[:i], jobQueue[i+1:]...)
				log.Infof("[PLC_Job] Job을 로봇에 배정했습니다. Job=%v, RobotId=%v", job.description, robot.id)
				break
			}
		}
	}

}

// getRobot
//
// 특정 상태의 로봇을 job에 배정하기 위한 함수
//
// - robotStatus: 배정받고자 하는 로봇의 상태 조건
func getRobot(robotStatus robotStatus, jobDescription string) (*robot, error) {
	job := newJob(robotStatus, jobDescription)
	jobQueue = append(jobQueue, job)
	go DistributeJob()

	// 로봇 - job 배정 대기
	robot := <-job.robotWaiting
	return robot, nil
}

// JobServeEmptyTrayToTable
//
// 빈 트레이 테이블로 서빙
//
// - slot: 빈 트레이가 있는 슬롯 정보
func JobServeEmptyTrayToTable(slot model.Slot) error {
	log.Infof("[PLC_로봇_Job] 빈 트레이 테이블로 서빙. slotId=%v", slot.SlotId)
	robot, err := getRobot(robotStatusAvailable, "빈 트레이 서빙")
	if err != nil {
		return err
	}

	robot.changeStatus(robotStatusWorking)
	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	// 슬롯으로 이동 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pullFromSlot(slot); err != nil {
		return err
	}
	// 트레이 꺼내기 완료 확인
	CheckCompletePlc("complete")

	resource.ReserveTable()
	if err := robot.moveToTable(); err != nil {
		return err
	}
	resource.ReleaseSlot(slot.SlotId)

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	// 트레이로 이동 완료 확인 & 뒷문 열림 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pushToTable(); err != nil {
		return err
	}

	// 테이블 점유를 해제하지 않고 대기상태 진입
	robot.changeStatus(robotStatusWaiting)

	return nil
}

// JobRetrieveEmptyTrayFromTable
//
// 테이블의 빈 트레이를 회수. robotStatusWaiting 상태의 로봇이 있으면 해당 로봇에게 job 요청함.
//
// - slot: 빈 트레이를 격납할 슬롯
func JobRetrieveEmptyTrayFromTable(slot model.Slot) error {
	log.Infof("[PLC_로봇_Job] 테이블의 빈 트레이를 회수. slotId=%v", slot.SlotId)

	var robot *robot

	// 대기 중인 로봇에게 job 우선 배정
	// waiting 로봇은 테이블을 점유하고 있으므로 resource.ReserveTable() 생략
	for _, r := range robots {
		if r.status == robotStatusWaiting {
			r, err := getRobot(robotStatusWaiting, "빈 트레이 회수")
			if err != nil {
				return err
			}
			robot = r
			break
		}
	}

	if robot == nil {
		r, err := getRobot(robotStatusAvailable, "빈 트레이 회수")
		if err != nil {
			return err
		}
		robot = r
		resource.ReserveTable()
		if err := robot.moveToTable(); err != nil {
			return err
		}
	}

	robot.changeStatus(robotStatusWorking)

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	// 테이블로 이동 완료 확인 & 뒷문 열림 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pullFromTable(); err != nil {
		return err
	}
	// 트레이 꺼내기 완료 확인
	CheckCompletePlc("complete")

	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationUp); err != nil {
		return err
	}

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose); err != nil {
		return err
	}

	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	resource.ReleaseTable()

	// 슬롯으로 이동 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pushToSlot(slot); err != nil {
		return err
	}
	// 트레이 넣기 완료 확인
	CheckCompletePlc("complete")

	resource.ReleaseSlot(slot.SlotId)

	robot.changeStatus(robotStatusAvailable)

	return nil
}

// JobInputItem
//
// 테이블의 물건을 슬롯에 가져다 놓기.
// 테이블 앞 대기중인 로봇에게 Job이 배정되고,
// 대기 중인 로봇이 없으면 사용 가능한 로봇에게 배정됨.
//
// - slot: 물건을 수납할 슬롯
func JobInputItem(slot model.Slot) error {
	log.Infof("[PLC_로봇_Job] 테이블의 물건을 슬롯에 가져다 놓기. slotId=%v", slot.SlotId)
	var robot *robot

	resource.ReserveSlot(slot.SlotId)

	// 대기 중인 로봇에게 job 우선 배정
	// waiting 로봇은 테이블을 점유하고 있으므로 resource.ReserveTable() 생략
	for _, r := range robots {
		if r.status == robotStatusWaiting {
			r, err := getRobot(robotStatusWaiting, "물건 수납")
			if err != nil {
				return err
			}
			robot = r
			break
		}
	}

	if robot == nil {
		r, err := getRobot(robotStatusAvailable, "물건 수납")
		if err != nil {
			return err
		}
		robot = r
		resource.ReserveTable()
		if err := robot.moveToTable(); err != nil {
			return err
		}
	}

	robot.changeStatus(robotStatusWorking)

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	// 테이블 이동 완료 확인 & 뒷문 열림 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pullFromTable(); err != nil {
		return err
	}
	// 트레이 꺼내기 완료 확인
	CheckCompletePlc("complete")

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose); err != nil {
		return err
	}
	// 뒷문 닫힘 완료 확인
	CheckCompletePlc("complete")

	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationUp); err != nil {
		return err
	}
	// 트레이 버퍼 올리기 완료 확인
	CheckCompletePlc("complete")

	//resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	resource.ReleaseTable()

	// 슬롯으로 이동 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pushToSlot(slot); err != nil {
		return err
	}

	resource.ReleaseSlot(slot.SlotId)

	robot.changeStatus(robotStatusAvailable)

	return nil
}

// JobOutputItem
//
// 슬롯의 물건을 테이블로 서빙.
//
// - slot: 물건을 꺼낼 슬롯
func JobOutputItem(slot model.Slot) error {
	log.Infof("[PLC_로봇_Job] 불출 Job 시작. slotId=%v", slot.SlotId)
	robot, err := getRobot(robotStatusAvailable, "물건 불출")
	if err != nil {
		return err
	}

	robot.changeStatus(robotStatusWorking)
	resource.ReserveSlot(slot.SlotId)

	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	// 슬롯 이동 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pullFromSlot(slot); err != nil {
		return err
	}
	// 트레이 꺼내기 완료 확인
	CheckCompletePlc("complete")

	resource.ReserveTable()

	if err := robot.moveToTable(); err != nil {
		return err
	}

	resource.ReleaseSlot(slot.SlotId)

	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationDown); err != nil {
		return err
	}

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	// 테이블 이동 완료 확인, 트레이 버퍼 내리기 완료 확인, 뒷문 열림 완료 확인
	CheckCompletePlc("complete")
	if err := robot.pushToTable(); err != nil {
		return err
	}
	// 트레이 넣기 완료 확인
	CheckCompletePlc("complete")

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose); err != nil {
		return err
	}

	// 불출작업 후 입주민이 수령 또는 취소할 때까지 테이블 점유 및 대기
	robot.changeStatus(robotStatusWaiting)

	log.Infof("[PLC_로봇_Job] 불출 Job 완료. slotId=%v", slot.SlotId)

	return nil
}

// JobMoveTray
//
// 빈 트레이 옮기기(정리).
//
// - from: 빈 트레이가 있는 슬롯
// - to: 빈 트레이를 가져다 놓을 슬롯
func JobMoveTray(from, to model.Slot) error {
	log.Infof("[PLC_로봇_Job] 정리. from_slotId=%v, to_slotId=%v", from.SlotId, to.SlotId)
	robot, err := getRobot(robotStatusAvailable, "정리")
	if err != nil {
		return err
	}

	robot.changeStatus(robotStatusWorking)

	resource.ReserveSlot(from.SlotId)
	resource.ReserveSlot(to.SlotId)
	if err := robot.moveToSlot(from); err != nil {
		return err
	}
	// 슬롯 이동 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pullFromSlot(from); err != nil {
		return err
	}
	// 트레이 꺼내기 완료 확인
	CheckCompletePlc("complete")

	if err := robot.moveToSlot(to); err != nil {
		return err
	}

	resource.ReleaseSlot(from.SlotId)

	// 슬롯 이동 완료 확인
	CheckCompletePlc("complete")

	if err := robot.pushToSlot(to); err != nil {
		return err
	}

	resource.ReleaseSlot(to.SlotId)

	robot.changeStatus(robotStatusAvailable)

	return nil
}

// JobWaitAtTable
//
// 로봇 하나를 테이블 앞에 대기
func JobWaitAtTable() error {
	log.Infof("[PLC_로봇_Job] 로봇 하나를 테이블 앞에 대기")
	robot, err := getRobot(robotStatusAvailable, "테이블 앞 대기")
	if err != nil {
		return err
	}
	resource.ReserveTable()
	if err := robot.moveToTable(); err != nil {
		return err
	}

	// 테이블 점유 및 대기
	robot.changeStatus(robotStatusWaiting)

	return nil
}

// JobDismiss
//
// 테이블 앞 로봇 대기 상태 해제
func JobDismiss() error {
	log.Infof("[PLC_로봇] 대기 상태 해제")

	for _, robot := range robots {
		if robot.status == robotStatusWaiting {
			robot, err := getRobot(robotStatusWaiting, "대기 해제")
			if err != nil {
				return err
			}
			resource.ReleaseTable()
			robot.changeStatus(robotStatusAvailable)
		}
	}

	return nil
}

// 업무 완료 확인
//
// data : 기대하는 값		***수정	// address , data2
func CheckCompletePlc(data interface{}) error {
	RespPlc = "waiting"
	go func() {
		time.Sleep(2 * time.Second)
		RespPlc = "complete"
	}()

	for {
		// 어떤 데이터를 가져올지 매개변수 추가
		log.Infof("[PLC] 10ms 마다 데이터 조회 중") // 조회한 데이터 struct에 저장

		if RespPlc == data {
			log.Info("업무 완료 응답")
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

// 트러블 감지
func SenseTrouble() {
	var waiting = make(chan string)
	log.Infof("[PLC] 10ms 마다 데이터 조회 중") // 조회한 데이터 struct에 저장

	/* if state.D0 == 1 {
		waiting <- "??"
	} */

	/* go func() {
		fmt.Println("실행")

		waiting <- "화재"
	}() */

	data := <-waiting
	switch data {
	case "화재":
		log.Infof("[PLC] 화재발생")
		// TODO - 사업자에게 알림
		// 키오스크 화면 변경
		return
	case "물품 끼임":
		log.Infof("[PLC] 물품 끼임")
		// TODO - 사업자에게 알림
		// 키오스크 화면 변경
		return
	case "물품 낙하":
		log.Infof("[PLC] 물품 낙하")
		// TODO - 사업자에게 알림
		// 키오스크 화면 변경
		return
	case "이물질 감지":
		log.Infof("[PLC] 이물질 감지")
		// TODO - 사업자에게 알림
		// 키오스크 화면 변경
		return

	}
}

/* // PLC
type PLC struct {
	addr string // 주소
	conn net.Conn
}

type State struct{
	D0 int
	D1 int

}

// 새로운 PLC를 생성
func NewPLC(addr string) (*PLC, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &PLC{
		addr: addr,
		conn: conn,
	}, nil
}

func (plc *PLC) Close() {
	plc.conn.Close()
}

// PLC로부터 응답읽어옴
func (plc *PLC) ReadRequest() ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := plc.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// PLC에 요청 보냄
func (plc *PLC) WriteRequest(req []byte) error {
	_, err := plc.conn.Write(req)
	return err
}

// PLC로부터 택배함의 상태를 조회
func (plc *PLC) Poll() ([]byte, error) {
	req := []byte{0x01}
	err := plc.WriteRequest(req)
	if err != nil {
		return nil, err
	}
	return plc.ReadRequest()
}


func InitConnPlc() {
	plc, err := NewPLC("192.168.1.100:502")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer plc.Close()

	// 10ms마다 PLC로부터 택배함의 상태를 조회
	for {
		req, err := plc.Poll()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(req)
		state := state{
			D0: req.D0,
			D1: req.D1,
		}

		time.Sleep(10 * time.Millisecond)

	}
}
*/
