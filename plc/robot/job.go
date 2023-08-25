package robot

import (
	"apcs_refactored/model"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/resource"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
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
	timestamp time.Time
}

var (
	jobQueue []*job
)

func newJob(robotStatus robotStatus) *job {
	job := &job{
		id:                  uuid.NewString(),
		requiredRobotStatus: robotStatus,
		robotWaiting:        make(chan *robot),
		timestamp:           time.Now(),
	}

	return job
}

func changeRobotStatus(robot *robot, robotStatus robotStatus) {
	robot.status = robotStatus
	go DistributeJob()
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

	job := jobQueue[0]

	for _, robot := range robots {
		if robot.status == job.requiredRobotStatus && job.robot == nil {
			job.robot = robot
			job.robotWaiting <- robot
			jobQueue = jobQueue[1:]
			log.Infof("[PLC_Job] Job을 로봇에 배정했습니다. Job: %v, Robot: %v", *job, *robot)
			break
		}
	}
}

// getRobot
//
// 특정 상태의 로봇을 job에 배정하기 위한 함수
//
// - robotStatus: 배정받고자 하는 로봇의 상태 조건
func getRobot(robotStatus robotStatus) (*robot, error) {
	job := newJob(robotStatus)
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
	log.Infof("[PLC_Robot_Job] 빈 트레이 테이블로 서빙. slot: %v", slot)
	robot, err := getRobot(available)
	if err != nil {
		return err
	}

	changeRobotStatus(robot, working)

	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	if err := robot.pullFromSlot(slot); err != nil {
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
	if err := robot.pushToTable(); err != nil {
		return err
	}
	resource.ReleaseTable()

	changeRobotStatus(robot, waiting)

	return nil
}

// JobRetrieveEmptyTrayFromTable
//
// 테이블의 빈 트레이를 회수. waiting 상태의 로봇이 있으면 해당 로봇에게 job 요청함.
//
// - slot: 빈 트레이를 격납할 슬롯
func JobRetrieveEmptyTrayFromTable(slot model.Slot) error {
	log.Infof("[PLC_Robot_Job] 테이블의 빈 트레이를 회수. slot: %v", slot)

	var robot *robot

	// 대기 중인 로봇에게 job 우선 배정
	waitingRobotExists := false
	for _, r := range robots {
		if r.status == waiting {
			waitingRobotExists = true
		}
	}

	if waitingRobotExists {
		r, err := getRobot(waiting)
		if err != nil {
			return err
		}
		robot = r
	} else {
		r, err := getRobot(available)
		if err != nil {
			return err
		}
		robot = r
	}

	changeRobotStatus(robot, working)

	resource.ReserveTable()
	if err := robot.moveToTable(); err != nil {
		return err
	}
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	if err := robot.pullFromTable(); err != nil {
		return err
	}
	resource.ReleaseTable()

	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	if err := robot.pushToSlot(slot); err != nil {
		return err
	}
	resource.ReleaseSlot(slot.SlotId)

	changeRobotStatus(robot, available)

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
	log.Infof("[PLC_Robot_Job] 테이블의 물건을 슬롯에 가져다 놓기. slot: %v", slot)

	var robot *robot

	// 대기 중인 로봇에게 job 우선 배정
	waitingRobotExists := false
	for _, r := range robots {
		if r.status == waiting {
			waitingRobotExists = true
		}
	}

	if waitingRobotExists {
		r, err := getRobot(waiting)
		if err != nil {
			return err
		}
		robot = r
	} else {
		r, err := getRobot(available)
		if err != nil {
			return err
		}
		robot = r
	}

	changeRobotStatus(robot, working)

	resource.ReserveTable()
	if err := robot.moveToTable(); err != nil {
		return err
	}
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen); err != nil {
		return err
	}
	if err := robot.pullFromTable(); err != nil {
		return err
	}
	resource.ReleaseTable()

	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	if err := robot.pushToSlot(slot); err != nil {
		return err
	}
	resource.ReleaseSlot(slot.SlotId)

	changeRobotStatus(robot, available)

	return nil
}

// JobOutputItem
//
// 슬롯의 물건을 테이블로 서빙.
//
// - slot: 물건을 꺼낼 슬롯
func JobOutputItem(slot model.Slot) error {
	log.Infof("[PLC_Robot_Job] 슬롯의 물건을 테이블로 서빙 시작. slot: %v", slot)
	robot, err := getRobot(available)
	if err != nil {
		return err
	}

	changeRobotStatus(robot, working)

	resource.ReserveSlot(slot.SlotId)
	if err := robot.moveToSlot(slot); err != nil {
		return err
	}
	if err := robot.pullFromSlot(slot); err != nil {
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
	if err := robot.pushToTable(); err != nil {
		return err
	}

	// 불출작업 후 입주민이 수령/취소할 때까지 보류
	changeRobotStatus(robot, waiting)

	log.Infof("[PLC_Robot_Job] 슬롯의 물건을 테이블로 서빙 완료. slot: %v", slot)

	return nil
}

// JobMoveTray
//
// 빈 트레이 옮기기(정리).
//
// - from: 빈 트레이가 있는 슬롯
// - to: 빈 트레이를 가져다 놓을 슬롯
func JobMoveTray(from, to model.Slot) error {
	log.Infof("[PLC_Robot_Job] 정리. slot_from: %v, slot_to: %v", from, to)
	robot, err := getRobot(available)
	if err != nil {
		return err
	}

	changeRobotStatus(robot, working)

	resource.ReserveSlot(from.SlotId)
	if err := robot.moveToSlot(from); err != nil {
		return err
	}
	if err := robot.pullFromSlot(from); err != nil {
		return err
	}
	resource.ReleaseSlot(from.SlotId)

	resource.ReserveSlot(to.SlotId)
	if err := robot.moveToSlot(to); err != nil {
		return err
	}
	if err := robot.pushToSlot(to); err != nil {
		return err
	}
	resource.ReleaseSlot(to.SlotId)

	changeRobotStatus(robot, available)

	return nil
}

// JobWaitAtTable
//
// 로봇 하나를 테이블 앞에 대기
func JobWaitAtTable() error {
	log.Infof("[PLC_Robot_Job] 로봇 하나를 테이블 앞에 대기")
	robot, err := getRobot(available)
	if err != nil {
		return err
	}

	resource.ReserveTable()
	if err := robot.moveToTable(); err != nil {
		return err
	}

	changeRobotStatus(robot, waiting)

	return nil
}

// JobDismiss
//
// 테이블 앞 로봇 대기 상태 해제
func JobDismiss() error {
	log.Infof("[PLC_로봇] 대기 상태 해제")

	robot, err := getRobot(waiting)
	if err != nil {
		return err
	}

	changeRobotStatus(robot, available)

	resource.ReleaseTable()

	return nil
}
