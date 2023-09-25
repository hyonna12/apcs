package resource

import (
	"apcs_refactored/config"
	"time"

	log "github.com/sirupsen/logrus"
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
// 슬롯에서 물건을 꺼낸 후 점유를 해제하기 때문에 이론 상 데드락이 발생할 경우는 없을 것으로 보이긴 함.
func CheckResolveDeadlock() {
	log.Infof("[PLC_resource] CheckResolveDeadlock()")
}

func ReserveTable() {

	if !tableReserved {
		tableReserved = true
		log.Infof("[PLC] 테이블 점유")

		return
	}

	// 테이블이 점유중인 경우 대기 채널을 등록하고 blocking
	waiting := make(chan struct{})
	tableWaitingQueue = append(tableWaitingQueue, waiting)
	for {
		select {
		case <-waiting:
			tableReserved = true
			log.Infof("[PLC] 테이블 점유")

			return
		// 일정 시간마다 데드락 확인 및 해결
		case <-time.After(deadlockCheckPeriod * time.Second):
			log.Warn("[PLC_resource] (table) 데드락 확인 및 해결 요청")
			go CheckResolveDeadlock()
		}
	}
}

func ReserveSlot(slotId int64) {

	isSlotReserved := slotReservationMap[slotId]
	if !isSlotReserved {
		slotReservationMap[slotId] = true
		log.Infof("[PLC] 슬롯 점유")

		return
	}

	// 슬롯이 점유중인 경우 대기 채널을 등록하고 blocking
	waiting := make(chan struct{})
	slotWaitingQueueMap[slotId] = append(slotWaitingQueueMap[slotId], waiting)
	for {
		select {
		case <-waiting:
			slotReservationMap[slotId] = true
			log.Infof("[PLC] 슬롯 점유")

			return
		// 일정 시간 마다 데드락 확인 및 해결
		case <-time.After(deadlockCheckPeriod * time.Second):
			log.Warn("[PLC_resource] (slot) 데드락 확인 및 해결 요청")
			go CheckResolveDeadlock()
		}
	}
}

func ReleaseTable() {
	log.Infof("[PLC] 테이블 점유 해제")

	tableReserved = false
	if len(tableWaitingQueue) > 0 {
		tableWaitingQueue[0] <- struct{}{}
		tableWaitingQueue = tableWaitingQueue[1:]
	}
}
func ReleaseSlot(slotId int64) {
	log.Infof("[PLC] 슬롯 점유 해제")

	slotReservationMap[slotId] = false
	if len(slotWaitingQueueMap[slotId]) > 0 {
		slotWaitingQueueMap[slotId][0] <- struct{}{}
		slotWaitingQueueMap[slotId] = slotWaitingQueueMap[slotId][1:]
	}
}
