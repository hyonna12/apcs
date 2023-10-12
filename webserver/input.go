package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/robot"
	"apcs_refactored/plc/trayBuffer"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Data struct {
	Robot Robot `json:"robot"`
	Item  Item  `json:"item"`
}
type Robot struct {
	X string `json:"x"`
	Z string `json:"z"`
}
type Item struct {
	Height string `json:"height"`
	Weight string `json:"weight"`
}

func DeliveryCompanyList(w http.ResponseWriter, r *http.Request) {
	// 택배회사 리스트 조회
	deliveryList, err := model.SelectDeliveryCompanyList()
	if err != nil {
		// changeKioskView
		// return
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// **삭제
	robot.SenseTrouble()

	Response(w, deliveryList, http.StatusOK, nil)
}

// CheckAdress
//
// [API] 배송정보 입력 화면에서 입력완료 버튼을 누른 경우 호출
func CheckAddress(w http.ResponseWriter, r *http.Request) {
	inputInfoRequest := InputInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(&inputInfoRequest)
	if err != nil {
		log.Error(err)
		Response(w, nil, http.StatusInternalServerError, err)
		return
	}

	if inputInfoRequest.Address == "" || inputInfoRequest.DeliveryId == "" || inputInfoRequest.TrackingNum == "" {
		log.Error(err)
		Response(w, nil, http.StatusBadRequest, errors.New("파라미터가 누락되었습니다"))
		return
	}

	ownerId, err := model.SelectOwnerIdByAddress(inputInfoRequest.Address)
	if ownerId == 0 {
		log.Error(err)
		Response(w, nil, http.StatusBadRequest, errors.New("입력하신 주소가 존재하지 않습니다"))
		return
	}
	if err != nil {
		// changeKioskView
		// return
		log.Error(err)
		Response(w, nil, http.StatusInternalServerError, err)
		return
	}
	Response(w, ownerId, http.StatusOK, nil)

}

// DeliveryInfoRequested
//
// [API] 배송정보 입력 화면에서 입력완료 버튼을 누른 경우 호출
//
// 성공 시 /input/input_item 호출
func DeliveryInfoRequested(w http.ResponseWriter, r *http.Request) {
	inputInfoRequest := InputInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(&inputInfoRequest)
	if err != nil {
		log.Error(err)
		Response(w, nil, http.StatusInternalServerError, err)
		return
	}

	tracking_num, _ := strconv.Atoi(inputInfoRequest.TrackingNum)
	itemDimension.TrackingNum = tracking_num

	// 버퍼에 빈트레이 유무 확인
	if !trayBuffer.Buffer.IsEmpty() {
		// 버퍼의 맨 위 트레이 사용
		trayBuffer.Buffer.Get()
		trayId := trayBuffer.Buffer.Peek().(int64)
		plc.TrayIdOnTable.Int64 = trayId
		log.Infof("[웹 핸들러] 테이블에 빈 트레이가 있어 사용. trayId=%v", trayId)

		err = plc.StandbyRobotAtTable()
		if err != nil {
			// changeKioskView
			// return
			log.Error(err)
			Response(w, nil, http.StatusInternalServerError, err)
		}
	} else {
		// 빈트레이를 가져올 슬롯 선택
		slotsWithEmptyTray, err := model.SelectSlotListWithEmptyTray()
		if len(slotsWithEmptyTray) == 0 {
			log.Info("[웹 핸들러] 빈 트레이가 존재하지 않음")
			Response(w, nil, http.StatusBadRequest, errors.New("빈 트레이가 존재하지 않습니다"))
			return
		}
		slotWithEmptyTray := slotsWithEmptyTray[0]
		trayId := slotWithEmptyTray.TrayId.Int64
		//** 수정
		log.Infof("[웹 핸들러] 빈 트레이를 가져올 slotId=%v, trayId=%v", slotWithEmptyTray.SlotId, trayId)

		if err != nil {
			log.Error(err)
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}

		err = plc.ServeEmptyTrayToTable(slotWithEmptyTray)
		if err != nil {
			log.Error(err)
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}

		// 버퍼에 트레이 추가
		trayBuffer.Buffer.Push(trayId)
		num := trayBuffer.Buffer.Count()
		model.InsertBufferState(num)
		plc.TrayIdOnTable.Int64 = trayId

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
	}

	err = plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationOpen)
	if err != nil {
		// changeKioskView
		// return
		Response(w, nil, http.StatusInternalServerError, err)
	}

	ownerId, err := model.SelectOwnerIdByAddress(inputInfoRequest.Address)
	if err != nil {
		// changeKioskView
		// return
		log.Error(err)
		Response(w, nil, http.StatusInternalServerError, err)
		return
	}

	log.Infof("[웹 핸들러] OwnerId=%v", ownerId)
	ownerIdStr := strconv.FormatInt(ownerId, 10)
	redirectUrl := "/input/input_item?deliveryId=" + inputInfoRequest.DeliveryId + "&ownerId=" + ownerIdStr
	Response(w, redirectUrl, http.StatusOK, nil)
}

// ItemSubmitted
//
// [API] 택배기사가 물건을 테이블에 올려놓은 후 호출
func ItemSubmitted(w http.ResponseWriter, r *http.Request) {

	deliveryIdStr = r.URL.Query().Get("deliveryId")
	ownerIdStr = r.URL.Query().Get("ownerId")

	err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose)
	if err != nil {
		// changeKioskView
		// return
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// TODO - temp
	//var itemDimension plc.ItemDimension
	// **수정
	// 물품 크기, 무게, 송장번호 조회
	item, err := plc.SenseItemInfo()
	itemDimension.Height = item.Height
	itemDimension.Width = item.Width
	itemDimension.Weight = item.Weight

	if err != nil {
		// changeKioskView
		// return
		Response(w, nil, http.StatusInternalServerError, err)
	}

	log.Printf("[제어서버] 아이템 크기/무게: %v", itemDimension)

	// 물품의 크기, 무게가 기준 초과되면 입고 취소
	if itemDimension.Height > 270 { // 단위:mm
		Response(w, nil, http.StatusBadRequest, errors.New("허용 높이 초과"))
		return
	}
	if itemDimension.Width > 10 {
		Response(w, nil, http.StatusBadRequest, errors.New("허용 너비 초과"))
		return
	}
	if itemDimension.Weight > 10 {
		Response(w, nil, http.StatusBadRequest, errors.New("허용 무게 초과"))
		return
	}

	// 물품을 수납할 최적 슬롯 찾기 **수정
	data := Data{Robot: Robot{X: "10", Z: "1"}, Item: Item{Height: strconv.Itoa(itemDimension.Height), Weight: strconv.Itoa(itemDimension.Weight)}}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	resp, err := http.Post("http://localhost:8080/get/best_slot", "application/json", buff)

	//var bestSlot model.Slot

	if err != nil {
		// 에러나면 직접 수납슬롯 구하기
		log.Error(err)
		slotList, err := model.SelectAvailableSlotList(itemDimension.Height)
		if len(slotList) == 0 {
			Response(w, nil, http.StatusBadRequest, errors.New("수납가능한 슬롯이 존재하지 않습니다"))
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

		bestSlot = slotList[0]
		log.Infof("[웹핸들러] 최적수납슬롯: slotId=%v", bestSlot.SlotId)

	} else {
		defer resp.Body.Close()

		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
		}
		json.Unmarshal(respData, &bestSlot)
		log.Infof("[웹핸들러] 최적 슬롯: slotId=%v", bestSlot.SlotId)
		if bestSlot.Lane == 0 {
			Response(w, nil, http.StatusBadRequest, errors.New("수납가능한 슬롯이 없습니다"))
			return
		}
	}
	Response(w, "/input/complete_input_item", http.StatusOK, nil)
}

func Input(w http.ResponseWriter, r *http.Request) {

	// 버퍼에서 사용한 트레이 삭제
	trayId := trayBuffer.Buffer.Pop().(int64)
	num := trayBuffer.Buffer.Count()
	model.InsertBufferState(num)

	value := trayBuffer.Buffer.Peek()

	if value == nil {
		log.Error("버퍼에 빈 트레이가 존재하지 않음")
		// TODO - 관리자에게 알림
	} else {
		id := value.(int64)
		plc.TrayIdOnTable.Int64 = id
	}

	// 송장번호, 물품높이, 택배기사, 수령인 정보 itemCreateRequest 에 넣어서 물품 db업데이트
	deliveryId, err := strconv.ParseInt(deliveryIdStr, 10, 64)
	if err != nil {
		// TODO - 에러처리
		log.Error(err)
		return
	}
	ownerId, err := strconv.ParseInt(ownerIdStr, 10, 64)
	if err != nil {
		// TODO - 에러처리
		log.Error(err)
		return
	}

	itemCreateRequest := model.ItemCreateRequest{
		ItemHeight:     itemDimension.Height,
		TrackingNumber: itemDimension.TrackingNum,
		DeliveryId:     deliveryId,
		OwnerId:        ownerId,
	}

	// 트랜잭션
	tx, err := model.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	itemId, err := model.InsertItem(itemCreateRequest, tx)
	if err != nil {
		// changeKioskView
		// return
		// TODO - 에러처리
		log.Error(err)
		return
	}

	// 슬롯, 트레이 db 업데이트
	// 트레이 아이디 추가
	trayUpdateRequest := model.TrayUpdateRequest{
		TrayOccupied: true,
		ItemId:       sql.NullInt64{Int64: itemId, Valid: true},
	}
	_, err = model.UpdateTray(trayId, trayUpdateRequest, tx)
	if err != nil {
		// TODO - 에러처리
		log.Error(err)
		return
	}

	// 아이템이 수납된 lane 슬롯 업데이트
	slots, err := model.SelectSlotListByLane(bestSlot.Lane)
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
			slot.ItemId = sql.NullInt64{Int64: itemId, Valid: true}
			slot.TrayId = sql.NullInt64{Int64: trayId, Valid: true}
			continue
		}

		height := float64(itemCreateRequest.ItemHeight)
		float := math.Ceil(height / 45)
		slotKeepCnt := int(float)

		// 물건이 차지하는 슬롯 갱신
		itemTopFloor := bestSlot.Floor - slotKeepCnt + 1

		if itemTopFloor <= slot.Floor && slot.Floor <= bestSlot.Floor {
			slot.SlotEnabled = false
			slot.SlotKeepCnt = 0
			slot.ItemId = sql.NullInt64{Int64: itemId, Valid: true}
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

	// 아이템 수납
	log.Infof("[웹핸들러] 아이템 수납")

	err = plc.InputItem(bestSlot)
	if err != nil {
		log.Error(err)
		// changeKioskView
		// return
	}

}

type StopRequest struct {
	Step string `json:"step"`
}

func StopInput(w http.ResponseWriter, r *http.Request) {
	log.Info("입고취소")
	stopRequest := StopRequest{}
	err := json.NewDecoder(r.Body).Decode(&stopRequest)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	if stopRequest.Step >= "2" {

		// 앞문이 열려있는지 확인한 후에 front open
		err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationOpen)
		if err != nil {
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
		}
		// 앞문 열림 완료 확인
		robot.CheckCompletePlc("complete")

		if err != nil {
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
		}

		// 물품 감지
		isItemOnTable, err := plc.SenseTableForItem()
		if err != nil {
			// changeKioskView
			// return
			Response(w, nil, http.StatusInternalServerError, err)
		}

		// 물품이 없다면(회수했다면) 앞문 닫기
		if !isItemOnTable {
			err = plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose)
			if err != nil {
				// changeKioskView
				// return
				Response(w, nil, http.StatusInternalServerError, err)
			}
			robot.JobDismiss()
		}
		boolStr := strconv.FormatBool(isItemOnTable)
		Response(w, boolStr, http.StatusOK, nil)
		return
	}

	if stopRequest.Step >= "1" {
		//robot.JobDismiss()
		Response(w, "OK", http.StatusOK, nil)
	}
}
