package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/robot"
	"apcs_refactored/plc/trayBuffer"
	"context"
	"database/sql"
	"errors"
	"math"
	"math/rand"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"
)

func SortItem(w http.ResponseWriter, r *http.Request) {
	log.Info("[PLC] 물품 정리")

	// 정리할 물품 선정 // **제거
	itemList, err := model.SelectSortItemList()
	if len(itemList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("정리가능한 물품이 존재하지 않습니다"))
		return
	}
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	num := rand.Intn(len(itemList) - 1)
	item := itemList[num]
	// 물품의 현재 슬롯
	currentSlot, err := model.SelectSlotByItemId(item.ItemId)
	log.Infof("[웹핸들러] 현재수납슬롯: slotId=%v", currentSlot.SlotId)

	if len(itemList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("해당 물품이 보관되어 있지 않습니다"))
		return
	}
	if err != nil {
		// changeKioskView
		// return
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 이동할 슬롯 선정 // **제거
	slotList, err := model.SelectAvailableSlotList(item.ItemHeight)
	if len(slotList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("이동가능한 슬롯이 존재하지 않습니다"))
		return
	}
	if err != nil {
		// changeKioskView
		// return
		Response(w, nil, http.StatusInternalServerError, err)
	}

	sort.SliceStable(slotList, func(i, j int) bool {
		return slotList[i].TransportDistance < slotList[j].TransportDistance
	})
	bestSlot := slotList[0]
	log.Infof("[웹핸들러] 최적수납슬롯: slotId=%v", bestSlot.SlotId)

	// 트레이 이동
	err = plc.MoveTray(currentSlot, bestSlot)
	if err != nil {
		// changeKioskView
		// return
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 트랜잭션
	tx, err := model.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// db 변경
	slots, err := model.SelectSlotListByLaneAndItemId(item.ItemId)
	if err != nil {
		log.Error(err)
		// TODO - DB 에러 처리
	}

	// 물건이 차지하던 슬롯 초기화
	for idx := range slots {
		slot := &slots[idx]

		if slot.ItemId.Int64 == item.ItemId {
			slot.SlotEnabled = true
			slot.ItemId = sql.NullInt64{Valid: false} // set null
			slot.TrayId = sql.NullInt64{Valid: false} // set null
		}
	}
	// slot-keep-cnt 갱신
	for idx := range slots {
		slot := &slots[idx]

		// 비어있는 슬롯에 대해서만 진행
		if !slot.SlotEnabled {
			continue
		}

		if idx == 0 { // 맨 위쪽 빈 슬롯인 경우
			slot.SlotKeepCnt = 1
		} else {
			slot.SlotKeepCnt = slots[idx-1].SlotKeepCnt + 1
		}
	}

	_, err = model.UpdateSlots(slots, tx)
	if err != nil {
		log.Error(err)
		// TODO - DB 에러 처리
	}

	// 아이템이 수납된 lane 슬롯 업데이트
	slots, err = model.SelectSlotListByLane(bestSlot.Lane)
	if err != nil {
		log.Error(err)
		// TODO - DB 에러 처리
		return
	}

	for idx := range slots {
		slot := &slots[idx]

		// 물건 가장 아랫부분 슬롯 갱신
		if slot.SlotId == bestSlot.SlotId {
			slot.SlotEnabled = false
			slot.SlotKeepCnt = 0
			slot.ItemId = sql.NullInt64{Int64: item.ItemId, Valid: true}
			slot.TrayId = currentSlot.TrayId
			continue
		}

		// 물건이 차지하는 슬롯 갱신
		height := float64(item.ItemHeight)
		float := math.Ceil(height / 45)
		slotKeepCnt := int(float)

		itemTopFloor := bestSlot.Floor - slotKeepCnt + 1
		if itemTopFloor <= slot.Floor && slot.Floor <= bestSlot.Floor {
			slot.SlotEnabled = false
			slot.SlotKeepCnt = 0
			slot.ItemId = sql.NullInt64{Int64: item.ItemId, Valid: true}
			slot.TrayId = sql.NullInt64{Valid: false} // set null
			continue
		}
	}

	// slot-keep-cnt 갱신
	for idx := range slots {
		slot := &slots[idx]

		// 비어있는 슬롯에 대해서만 진행
		if !slot.SlotEnabled {
			continue
		}

		if idx == 0 { // 맨 위쪽 빈 슬롯인 경우
			slot.SlotKeepCnt = 1
		} else {
			slot.SlotKeepCnt = slots[idx-1].SlotKeepCnt + 1
		}
	}

	_, err = model.UpdateSlots(slots, tx)
	if err != nil {
		log.Error(err)
		// TODO - DB 에러 처리
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	Response(w, "OK", http.StatusOK, nil)
}

func SortTrayBuffer(w http.ResponseWriter, r *http.Request) {
	log.Info("[PLC] 트레이 버퍼 정리")

	// 트레이 버퍼의 개수를 조회
	count := trayBuffer.Buffer.Count()

	// 15개면 정리 종료
	if count == 15 {
		Response(w, count, http.StatusOK, nil)
		return

	} else if count > 15 {
		// 15개 초과 시 회수.
		err := RetrieveEmptyTrayFromTableAndUpdateDb()
		if err != nil {
			log.Error(err)
			Response(w, nil, http.StatusBadRequest, errors.New(err.Error()))
			return
			// TODO - 에러 처리
		}

		trayBuffer.Buffer.Pop()
		num := trayBuffer.Buffer.Count()
		model.InsertBufferState(num)

		trayId := trayBuffer.Buffer.Peek().(int64)
		plc.TrayIdOnTable.Int64 = trayId

		Response(w, num, http.StatusOK, nil)
		return

	} else if count < 15 {
		// 15개 미만 채우기
		// 빈트레이를 가져올 슬롯 선택
		slotsWithEmptyTray, err := model.SelectSlotListWithEmptyTray()

		if len(slotsWithEmptyTray) == 0 {
			log.Info("[웹 핸들러] 빈 트레이가 존재하지 않음")
			Response(w, nil, http.StatusBadRequest, errors.New("빈 트레이가 존재하지 않습니다"))
			return
		}
		if err != nil {
			log.Error(err)
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}

		// TODO - 빈 슬롯 선정 최적화
		slotWithEmptyTray := slotsWithEmptyTray[0]
		trayId := slotWithEmptyTray.TrayId.Int64
		log.Infof("[웹 핸들러] 빈 트레이를 가져올 slotId=%v, trayId=%v", slotWithEmptyTray.SlotId, trayId)

		err = plc.SetUpTrayBuffer(trayBuffer.BufferOperationDown)
		if err != nil {
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
		}

		trayBuffer.Buffer.Push(trayId)
		num := trayBuffer.Buffer.Count()
		model.InsertBufferState(num)
		plc.TrayIdOnTable.Int64 = trayId

		err = plc.ServeEmptyTrayToTable(slotWithEmptyTray)
		if err != nil {
			log.Error(err)
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}
		robot.JobDismiss()

		slotUpdateRequest := model.SlotUpdateRequest{
			Lane:  slotWithEmptyTray.Lane,
			Floor: slotWithEmptyTray.Floor,
		}

		tx, err := model.DB.BeginTx(context.Background(), nil)
		if err != nil {
			return
		}
		defer func(tx *sql.Tx) {
			_ = tx.Rollback()
		}(tx)

		_, err = model.UpdateSlotToEmptyTray(slotUpdateRequest, tx)
		if err != nil {
			log.Error(err)
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}
		err = tx.Commit()
		if err != nil {
			return
		}
		err = plc.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose)
		if err != nil {
			log.Error(err)
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}
		Response(w, num, http.StatusOK, nil)
		return
	}
}
