package robot

import (
	"apcs_refactored/model"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/resource"
	"apcs_refactored/plc/trayBuffer"
	"time"

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
	if err := robot.pullTray(); err != nil {
		return err
	}
	resource.ReleaseSlot(slot.SlotId)

	resource.ReserveTable()
	if err := robot.moveToTable(); err != nil {
		return err
	}
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	if err := robot.pushTray(); err != nil {
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
	}

	robot.changeStatus(robotStatusWorking)

	if err := robot.moveToTable(); err != nil {
		return err
	}
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	if err := robot.pullTray(); err != nil {
		return err
	}
	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationUp); err != nil {
		return err
	}
	resource.ReleaseTable()

	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	if err := robot.pushTray(); err != nil {
		return err
	}
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
	}

	robot.changeStatus(robotStatusWorking)

	resp, err := GetRobotState()
	if err != nil {
		return err
	}
	if resp[robot.id-1].Location.Lane != 0 || resp[robot.id-1].Location.Floor != 0 {
		if err := robot.moveToTable(); err != nil {
			return err
		}
	}

	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	if err := robot.pullTray(); err != nil {
		return err
	}
	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationUp); err != nil {
		return err
	}

	resource.ReleaseTable()

	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	if err := robot.pushTray(); err != nil {
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
	if err := robot.pullTray(); err != nil {
		return err
	}
	resource.ReleaseSlot(slot.SlotId)

	resource.ReserveTable()

	if err := robot.moveToTable(); err != nil {
		return err
	}
	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationDown); err != nil {
		return err
	}
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	if err := robot.pushTray(); err != nil {
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
	if err := robot.moveToSlot(from); err != nil {
		return err
	}
	if err := robot.pullTray(); err != nil {
		return err
	}
	resource.ReleaseSlot(from.SlotId)

	resource.ReserveSlot(to.SlotId)
	if err := robot.moveToSlot(to); err != nil {
		return err
	}
	if err := robot.pushTray(); err != nil {
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
