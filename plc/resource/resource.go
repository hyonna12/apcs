package resource

import (
	"apcs_refactored/config"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	// tableReserved
	//
	// 테이블 점유 여부를 표시
	//   - true: 점유 중
	//   - false: 사용 가능
	tableReserved = false

	// tableWaitingQueue
	//
	// 테이블의 점유가 해제되기를 기다리는 채널 큐
	tableWaitingQueue = make([]chan struct{}, 0)

	// slotReservationMap
	//
	// 슬롯 점유 여부를 표시하는 맵
	// 슬롯 id를 key로 사용함
	//   - true: 점유 중
	//   - false: 사용 가능
	slotReservationMap = make(map[int64]bool)

	// slotWaitingQueueMap
	//
	// 슬롯의 점유가 해제되기를 기다리는 채널 큐 맵
	slotWaitingQueueMap = make(map[int64][]chan struct{})

	deadlockCheckPeriod time.Duration
)

func InitResources(slotIds []int64) {
	deadlockCheckPeriod = time.Duration(config.Config.Plc.Resource.DeadlockCheckPeriod)

	for _, slotId := range slotIds {
		slotWaitingQueueMap[slotId] = make([]chan struct{}, 0)
		slotReservationMap[slotId] = false
	}
}

// CheckResolveDeadlock
//
// TODO - 로봇 데드락 해결 로직
// 두 개의 로봇이 서로 점유하고 있는 자원을 기다릴 때 발생하는 데드락 감시 및 해결.
func CheckResolveDeadlock() {
	log.Infof("Checking Deadlock condition")
}

func ReserveTable() {
	if !tableReserved {
		tableReserved = true
		return
	}

	// 테이블이 점유중인 경우 대기 채널을 등록하고 blocking
	waiting := make(chan struct{})
	tableWaitingQueue = append(tableWaitingQueue, waiting)
	for {
		select {
		case <-waiting:
			tableReserved = true
			return
		// 일정 시간마다 데드락 확인 및 해결
		case <-time.After(deadlockCheckPeriod * time.Second):
			log.Warn("[PLC_resource] 데드락 확인 및 해결 요청")
			go CheckResolveDeadlock()
		}
	}
}

func ReserveSlot(slotId int64) {
	isSlotReserved := slotReservationMap[slotId]
	if !isSlotReserved {
		slotReservationMap[slotId] = true
		return
	}

	// 슬롯이 점유중인 경우 대기 채널을 등록하고 blocking
	waiting := make(chan struct{})
	slotWaitingQueueMap[slotId] = append(slotWaitingQueueMap[slotId], waiting)
	for {
		select {
		case <-waiting:
			slotReservationMap[slotId] = true
			return
		// 일정 시간마다 데드락 확인 및 해결
		case <-time.After(deadlockCheckPeriod * time.Second):
			log.Warn("[PLC_resource] 데드락 확인 및 해결 요청")
			go CheckResolveDeadlock()
		}
	}
}

func ReleaseTable() {
	tableReserved = false
	// TODO - temp
	if len(tableWaitingQueue) > 0 {
		tableWaitingQueue[0] <- struct{}{}
		tableWaitingQueue = tableWaitingQueue[1:]
	}
}
func ReleaseSlot(slotId int64) {
	slotReservationMap[slotId] = false
	if len(slotWaitingQueueMap[slotId]) > 0 {
		slotWaitingQueueMap[slotId][0] <- struct{}{}
		slotWaitingQueueMap[slotId] = slotWaitingQueueMap[slotId][1:]
	}
}
