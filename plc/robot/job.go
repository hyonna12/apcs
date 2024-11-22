package robot

import (
	"apcs_refactored/model"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/resource"
	"apcs_refactored/plc/trayBuffer"
	"fmt"
	"time"

	"github.com/google/uuid"

	"apcs_refactored/plc/conn"

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
	jobQueue          []*job
	RespPlc           interface{}
	outputAbortSignal = false
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

	// 슬롯으로 이동
	moveCommandId := GenerateCommandId()
	if err := robot.moveToSlot(slot, moveCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(moveCommandId); err != nil {
		return err
	}

	// 슬롯 방향으로 회전
	rotateCommandId := GenerateCommandId()
	if err := robot.rotateHandler(getDirectionFromSlotId(slot.SlotId), rotateCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateCommandId); err != nil {
		return err
	}

	// 트레이 꺼내기
	pullCommandId := GenerateCommandId()
	if err := robot.pullFromSlot(slot, pullCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pullCommandId); err != nil {
		return err
	}

	// 원위치로 회전
	rotateHomeCommandId := GenerateCommandId()
	if err := robot.rotateHandlerHome(rotateHomeCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateHomeCommandId); err != nil {
		return err
	}

	// 테이블로 이동
	resource.ReserveTable()
	moveTableCommandId := GenerateCommandId()
	if err := robot.moveToTable(moveTableCommandId); err != nil {
		return err
	}
	resource.ReleaseSlot(slot.SlotId)

	// 뒷문 열기
	doorOpenCommandId := GenerateCommandId()
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen, doorOpenCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(doorOpenCommandId); err != nil {
		return err
	}

	// 트레이 넣기
	pushCommandId := GenerateCommandId()
	if err := robot.pushToTable(pushCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pushCommandId); err != nil {
		return err
	}

	// 뒷문 닫기
	doorCloseCommandId := GenerateCommandId()
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose, doorCloseCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(doorCloseCommandId); err != nil {
		return err
	}

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
		moveTableCommandId := GenerateCommandId()
		if err := robot.moveToTable(moveTableCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(moveTableCommandId); err != nil {
			return err
		}
	}

	robot.changeStatus(robotStatusWorking)

	// 뒷문 열기
	doorOpenCommandId := GenerateCommandId()
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen, doorOpenCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(doorOpenCommandId); err != nil {
		return err
	}

	// 트레이 꺼내기
	pullCommandId := GenerateCommandId()
	if err := robot.pullFromTable(pullCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pullCommandId); err != nil {
		return err
	}

	// 트레이 버퍼 올리기
	bufferUpCommandId := GenerateCommandId()
	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationUp, bufferUpCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(bufferUpCommandId); err != nil {
		return err
	}

	// 뒷문 닫기
	doorCloseCommandId := GenerateCommandId()
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose, doorCloseCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(doorCloseCommandId); err != nil {
		return err
	}

	// 슬롯으로 이동
	resource.ReserveSlot(slot.SlotId)
	moveSlotCommandId := GenerateCommandId()
	if err := robot.moveToSlot(slot, moveSlotCommandId); err != nil {
		return err
	}
	resource.ReleaseTable()
	if err := CheckCompletePlc(moveSlotCommandId); err != nil {
		return err
	}

	// 슬롯 방향으로 회전
	rotateCommandId := GenerateCommandId()
	if err := robot.rotateHandler(getDirectionFromSlotId(slot.SlotId), rotateCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateCommandId); err != nil {
		return err
	}

	// 트레이 넣기
	pushCommandId := GenerateCommandId()
	if err := robot.pushToSlot(slot, pushCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pushCommandId); err != nil {
		return err
	}

	// 원위치로 회전
	rotateHomeCommandId := GenerateCommandId()
	if err := robot.rotateHandlerHome(rotateHomeCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateHomeCommandId); err != nil {
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
	log.Infof("[PLC_로봇_Job] 물품 수납 시작. slotId=%v", slot.SlotId)
	var robot *robot
	resource.ReserveSlot(slot.SlotId)

	// 대기 중인 로봇에게 job 우선 배정
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
		moveTableCommandId := GenerateCommandId()
		if err := robot.moveToTable(moveTableCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(moveTableCommandId); err != nil {
			return err
		}
	}

	robot.changeStatus(robotStatusWorking)

	// 뒷문 열기
	doorOpenCommandId := GenerateCommandId()
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen, doorOpenCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(doorOpenCommandId); err != nil {
		return err
	}

	// 트레이 꺼내기
	pullCommandId := GenerateCommandId()
	if err := robot.pullFromTable(pullCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pullCommandId); err != nil {
		return err
	}

	// 뒷문 닫기
	doorCloseCommandId := GenerateCommandId()
	if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose, doorCloseCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(doorCloseCommandId); err != nil {
		return err
	}

	// 트레이 버퍼 올리기
	bufferUpCommandId := GenerateCommandId()
	if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationUp, bufferUpCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(bufferUpCommandId); err != nil {
		return err
	}

	// 슬롯으로 이동
	moveSlotCommandId := GenerateCommandId()
	if err := robot.moveToSlot(slot, moveSlotCommandId); err != nil {
		return err
	}
	resource.ReleaseTable()
	if err := CheckCompletePlc(moveSlotCommandId); err != nil {
		return err
	}

	// 슬롯 방향으로 회전
	rotateCommandId := GenerateCommandId()
	if err := robot.rotateHandler(getDirectionFromSlotId(slot.SlotId), rotateCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateCommandId); err != nil {
		return err
	}

	// 트레이 넣기
	pushCommandId := GenerateCommandId()
	if err := robot.pushToSlot(slot, pushCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pushCommandId); err != nil {
		return err
	}

	// 원위치로 회전
	rotateHomeCommandId := GenerateCommandId()
	if err := robot.rotateHandlerHome(rotateHomeCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateHomeCommandId); err != nil {
		return err
	}

	resource.ReleaseSlot(slot.SlotId)
	robot.changeStatus(robotStatusAvailable)
	return nil

	if err := robot.completeJob(); err != nil {
		return err
	}

	return nil
}

var OutputRobots []*OutputRobotState

type OutputRobotState struct {
	RobotId int
	ItemId  int64
	SlotId  int64
}

func CountOutputRobotList() int {
	count := len(OutputRobots)
	return count
}
func DeleteOutputRobotList() {
	OutputRobots = OutputRobots[1:]
}

func GetOutputRobotList() []*OutputRobotState {
	return OutputRobots
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

	if !outputAbortSignal {
		resource.ReserveSlot(slot.SlotId)
		//slot.Status = "slotStatusGoing"

		// 슬롯으로 이동
		moveSlotCommandId := GenerateCommandId()
		if err := robot.moveToSlot(slot, moveSlotCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(moveSlotCommandId); err != nil {
			return err
		}

		// 슬롯 방향으로 회전
		rotateCommandId := GenerateCommandId()
		if err := robot.rotateHandler(getDirectionFromSlotId(slot.SlotId), rotateCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(rotateCommandId); err != nil {
			return err
		}

		// 트레이 꺼내기
		pullCommandId := GenerateCommandId()
		if err := robot.pullFromSlot(slot, pullCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(pullCommandId); err != nil {
			return err
		}

		// 원위치로 회전
		rotateHomeCommandId := GenerateCommandId()
		if err := robot.rotateHandlerHome(rotateHomeCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(rotateHomeCommandId); err != nil {
			return err
		}

		resource.ReserveTable()
		// 불출 전체 취소 여부 확인
		if outputAbortSignal {
			log.Infof("불출 취소 여부: %v", outputAbortSignal)
			resource.ReleaseTable()
			robot.changeStatus(robotStatusSlot)

			outputRobot := OutputRobotState{RobotId: robot.id, ItemId: slot.ItemId.Int64, SlotId: slot.SlotId}
			OutputRobots = append(OutputRobots, &outputRobot)
			log.Infof("불출 로봇 상태, %v", outputRobot)

			return nil

		} else if !outputAbortSignal {
			// 테이블로 이동
			moveTableCommandId := GenerateCommandId()
			if err := robot.moveToTable(moveTableCommandId); err != nil {
				return err
			}
			if err := CheckCompletePlc(moveTableCommandId); err != nil {
				return err
			}

			resource.ReleaseSlot(slot.SlotId)

			// 트레이 버퍼 내리기
			bufferDownCommandId := GenerateCommandId()
			if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationDown, bufferDownCommandId); err != nil {
				return err
			}
			if err := CheckCompletePlc(bufferDownCommandId); err != nil {
				return err
			}

			// 뒷문 열기
			doorOpenCommandId := GenerateCommandId()
			if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen, doorOpenCommandId); err != nil {
				return err
			}
			if err := CheckCompletePlc(doorOpenCommandId); err != nil {
				return err
			}

			// 트레이 넣기
			pushCommandId := GenerateCommandId()
			if err := robot.pushToTable(pushCommandId); err != nil {
				return err
			}
			if err := CheckCompletePlc(pushCommandId); err != nil {
				return err
			}

			// 뒷문 닫기
			doorCloseCommandId := GenerateCommandId()
			if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose, doorCloseCommandId); err != nil {
				return err
			}
			if err := CheckCompletePlc(doorCloseCommandId); err != nil {
				return err
			}

			robot.changeStatus(robotStatusWaiting)

			outputRobot := OutputRobotState{RobotId: robot.id, ItemId: slot.ItemId.Int64, SlotId: slot.SlotId}
			OutputRobots = append(OutputRobots, &outputRobot)
			log.Infof("불출 로봇 상태, %v", outputRobot)
			log.Infof("[PLC_로봇_Job] 불출 Job 완료. slotId=%v", slot.SlotId)
		}
	} else {
		// 불출 취소된 경우 대기 위치로 복귀
		returnHomeCommandId := GenerateCommandId()
		if err := robot.returnToHome(returnHomeCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(returnHomeCommandId); err != nil {
			return err
		}
		robot.changeStatus(robotStatusAvailable)
	}

	return nil
}

// JobReturnItem
//
// 테이블의 물건을 슬롯에 가져다 놓기.
// 테이블 앞 대기중인 로봇에게 Job이 배정되고,
// 대기 중인 로봇이 없으면 사용 가능한 로봇에게 배정됨.
//
// - slot: 물건을 수납할 슬롯
func JobReturnItem(slot model.Slot, robotId int) error {
	log.Infof("[PLC_로봇_Job] 불출중인 물건을 슬롯에 가져다 놓기. slotId=%v", slot.SlotId)
	var robot *robot
	isItem := false

	// 점유 상태 확인
	check := resource.CheckSlotReserve(slot.SlotId)
	// 슬롯 점유하고 있으면
	if check {
		robot, _ = getRobot(robotStatusSlot, "물건재수납")
		isItem = true
	} else {
		// 테이블에 놓여있으면
		robot, _ = getRobot(robotStatusWaiting, "물건재수납")
		resource.ReserveSlot(slot.SlotId)
	}

	robot.changeStatus(robotStatusWorking)

	// 로봇 팔에 물품이 있는지 확인
	if !isItem {
		// 뒷문 열기
		doorOpenCommandId := GenerateCommandId()
		if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationOpen, doorOpenCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(doorOpenCommandId); err != nil {
			return err
		}

		// 테이블에서 트레이 꺼내기
		pullCommandId := GenerateCommandId()
		if err := robot.pullFromTable(pullCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(pullCommandId); err != nil {
			return err
		}

		// 뒷문 닫기
		doorCloseCommandId := GenerateCommandId()
		if err := door.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose, doorCloseCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(doorCloseCommandId); err != nil {
			return err
		}

		// 트레이 버퍼 올리기
		bufferUpCommandId := GenerateCommandId()
		if err := trayBuffer.SetUpTrayBuffer(trayBuffer.BufferOperationUp, bufferUpCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(bufferUpCommandId); err != nil {
			return err
		}

		// 슬롯으로 이동
		moveSlotCommandId := GenerateCommandId()
		if err := robot.moveToSlot(slot, moveSlotCommandId); err != nil {
			return err
		}

		resource.ReleaseTable()
		if err := CheckCompletePlc(moveSlotCommandId); err != nil {
			return err
		}
	}

	// 슬롯 방향으로 회전
	rotateCommandId := GenerateCommandId()
	if err := robot.rotateHandler(getDirectionFromSlotId(slot.SlotId), rotateCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateCommandId); err != nil {
		return err
	}

	// 트레이 넣기
	pushCommandId := GenerateCommandId()
	if err := robot.pushToSlot(slot, pushCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pushCommandId); err != nil {
		return err
	}

	// 원위치로 회전
	rotateHomeCommandId := GenerateCommandId()
	if err := robot.rotateHandlerHome(rotateHomeCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateHomeCommandId); err != nil {
		return err
	}

	if err := robot.completeJob(); err != nil {
		return err
	}

	resource.ReleaseSlot(slot.SlotId)
	robot.changeStatus(robotStatusAvailable)

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

	// from 슬롯으로 이동
	moveFromCommandId := GenerateCommandId()
	if err := robot.moveToSlot(from, moveFromCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(moveFromCommandId); err != nil {
		return err
	}

	// from 슬롯 방향으로 회전
	rotateFromCommandId := GenerateCommandId()
	if err := robot.rotateHandler(getDirectionFromSlotId(from.SlotId), rotateFromCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateFromCommandId); err != nil {
		return err
	}

	// 트레이 꺼내기
	pullCommandId := GenerateCommandId()
	if err := robot.pullFromSlot(from, pullCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pullCommandId); err != nil {
		return err
	}

	// 원위치로 회전
	rotateHomeFromCommandId := GenerateCommandId()
	if err := robot.rotateHandlerHome(rotateHomeFromCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateHomeFromCommandId); err != nil {
		return err
	}

	// to 슬롯으로 이동
	moveToCommandId := GenerateCommandId()
	if err := robot.moveToSlot(to, moveToCommandId); err != nil {
		return err
	}
	resource.ReleaseSlot(from.SlotId)
	if err := CheckCompletePlc(moveToCommandId); err != nil {
		return err
	}

	// to 슬롯 방향으로 회전
	rotateToCommandId := GenerateCommandId()
	if err := robot.rotateHandler(getDirectionFromSlotId(to.SlotId), rotateToCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateToCommandId); err != nil {
		return err
	}

	// 트레이 넣기
	pushCommandId := GenerateCommandId()
	if err := robot.pushToSlot(to, pushCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(pushCommandId); err != nil {
		return err
	}

	// 원위치로 회전
	rotateHomeToCommandId := GenerateCommandId()
	if err := robot.rotateHandlerHome(rotateHomeToCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(rotateHomeToCommandId); err != nil {
		return err
	}

	resource.ReleaseSlot(to.SlotId)
	if err := robot.completeJob(); err != nil {
		return err
	}
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
	moveTableCommandId := GenerateCommandId()
	if err := robot.moveToTable(moveTableCommandId); err != nil {
		return err
	}
	if err := CheckCompletePlc(moveTableCommandId); err != nil {
		return err
	}

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

			returnHomeCommandId := GenerateCommandId()
			if err := robot.returnToHome(returnHomeCommandId); err != nil {
				return err
			}
			if err := CheckCompletePlc(returnHomeCommandId); err != nil {
				return err
			}

			robot.changeStatus(robotStatusAvailable)
		}
	}
	return nil
}

func ChangeOutPutAbortSignal(data bool) {
	//log.Infof("불출 취소 여부 변경:%v", data)
	outputAbortSignal = data
}

func CheckOutPutAbortSignal() bool {
	log.Infof("불출 취소 여부 확인:%v", outputAbortSignal)
	return outputAbortSignal
}

// CheckCompletePlc - PLC 명령 완료 대기
//
// commandId: 완료를 기다릴 명령의 ID
func CheckCompletePlc(commandId string) error {
	const timeout = 30 * time.Second
	start := time.Now()

	for {
		// 타임아웃 체크
		if time.Since(start) > timeout {
			return fmt.Errorf("작업 완료 대기 시간 초과 (%v초)", timeout.Seconds())
		}

		// PLC 서버로부터 응답 확인
		resp, err := conn.GetResponse()
		if err != nil {
			log.Errorf("[PLC] 응답 확인 실패: %v", err)
			return err
		}

		// commandId가 일치하고 성공적으로 완료되었는지 확인
		if resp.CommandId == commandId && resp.Success {
			log.Debugf("[PLC] 작업 완료 (commandId: %s)", commandId)
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// 명령 전송 시 고유 ID 생성
func GenerateCommandId() string {
	return uuid.NewString()
}

func (r *robot) completeJob() error {
	// 작업 완료 후 대기 위치로 복귀
	if r.status != robotStatusWaiting { // 대기 상태가 아닐 때만 복귀
		returnHomeCommandId := GenerateCommandId()
		if err := r.returnToHome(returnHomeCommandId); err != nil {
			return err
		}
		if err := CheckCompletePlc(returnHomeCommandId); err != nil {
			return err
		}
	}
	return nil
}

func getDirectionFromSlotId(slotId int64) string {
	if slotId <= 288 {
		return "rear" // 1~288은 홀수 레인(후면)
	}
	return "front" // 289~576은 짝수 레인(전면)
}
