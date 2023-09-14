package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/door"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

// CheckItemExists - [API] 동호수 입력 시 호출
func CheckItemExists(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	log.Infof("[불출] 입주민 주소 입력: %v", address)

	exists, err := model.SelectItemExistsByAddress(address)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	if exists {
		_, err = fmt.Fprint(w, fmt.Sprintf("/output/item_list?address=%v", address))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Item Not Found", http.StatusNotFound)
	}
}

// GetItemList - [API] 아이템 목록 반환
func GetItemList(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	address := r.URL.Query().Get("address")
	itemListResponses, err := model.SelectItemListByAddress(address)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	response := CommonResponse{
		Data:   itemListResponses,
		Status: http.StatusOK,
		Error:  nil,
	}

	data, _ := json.Marshal(response)
	_, err = fmt.Fprint(w, string(data))
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ItemOutputCheckPassword - [API] 비밀번호가 제출된 경우 호출
func ItemOutputCheckPassword(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	request := &struct {
		ItemId   int64 `json:"item_id"`
		Password int   `json:"password"`
	}{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// TODO - 비밀번호 해싱
	password, err := model.SelectPasswordByItemId(request.ItemId)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	if request.Password == password {
		Response(w, nil, http.StatusOK, nil)
	} else {
		Response(w, nil, http.StatusBadRequest, errors.New("잘못된 비밀번호입니다"))
	}
}

// ItemOutputReturn
//
// [API] "택배를 확인해 주세요" 화면에서 "반납" 버튼을 누른 경우 호출
// 비밀번호 입력 화면에서 "취소" 버튼을 누른 경우 호출
// "택배를 꺼내 주세요" 화면에서 5초 경과 후 호출
func ItemOutputReturn(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemId, err := strconv.ParseInt(r.URL.Query().Get("itemId"), 10, 64)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusBadRequest)
	}

	log.Infof("[웹핸들러] 물건 반납 요청 접수. itemId=%v", itemId)

	if err = plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose); err != nil {
		log.Error(err)
		// TODO - PLC 에러처리
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// 택배 반환
	// TODO - 반환하는 김에 정리(최적 슬롯 알고리즘) - 꺼낸 슬롯의 원래 자리는 비어있다고 가정하고 선정
	slot, err := model.SelectSlotByItemId(itemId)
	if err != nil {
		// TODO - DB 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	go func() {
		if _, err = plc.InputItem(slot); err != nil {
			// TODO - PLC 에러처리
			log.Error(err)
		}
	}()

	delete(requestList, itemId)

	if len(requestList) > 0 {
		// Request가 남아 있는 경우- 택배가 나오고 있습니다 화면으로
		log.Info("[웹핸들러] 불출 요청이 남아 있어 택배 나오는 중 화면 출력")

		err := ChangeKioskView("/output/ongoing")
		if err != nil {
			log.Error(err)
		}
	} else {
		// Request가 남아있지 않은 경우 - 택배 찾기가 취소되었습니다 화면으로
		log.Info("[웹핸들러] 불출 요청이 남아 있지 않아 감사합니다 화면 출력")

		err := ChangeKioskView("/output/cancel")
		if err != nil {
			log.Error(err)
		}
	}

	Response(w, nil, http.StatusOK, nil)
}

// TODO - 시간 초과에 의한 반납
// ItemOutputReturnByTimeout
//
// [API] "택배를 꺼내 주세요" 화면에서 5초 경과 후 호출
func ItemOutputReturnByTimeout(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	log.Info("[웹핸들러] 시간 초과에 의한 모든 반납 취소.")

	if err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose); err != nil {
		log.Error(err)
		// TODO - 앞문 닫힘 불가 시 에러처리
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	for _, request := range requestList {
		itemId := request.itemId

		// 택배 반환
		// TODO - 반환하는 김에 정리(최적 슬롯 알고리즘) - 꺼낸 슬롯의 원래 자리는 비어있다고 가정하고 선정
		slot, err := model.SelectSlotByItemId(itemId)
		if err != nil {
			// TODO - DB 에러처리
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		go func() {
			if _, err = plc.InputItem(slot); err != nil {
				// TODO - PLC 에러처리
				log.Error(err)
			}
		}()

		delete(requestList, itemId)
	}

	err := ChangeKioskView("/output/cancel")
	if err != nil {
		log.Error(err)
	}

	Response(w, nil, http.StatusOK, nil)
}

// ItemOutputTakeout
//
// TODO - temp - [API] 키오스크 물건 꺼내기 버튼 (시뮬레이션 용)
func ItemOutputTakeout(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	plc.IsItemOnTable = false
	Response(w, nil, http.StatusOK, nil)
}

// SenseTableForItem - [API] "택배를 꺼내주세요" 화면에서 매 초마다 호출
func SenseTableForItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	isItemOnTable, err := plc.SenseTableForItem()
	if err != nil {
		log.Error(err)
		// TODO - 에러처리
		Response(w, nil, http.StatusInternalServerError, nil)
	}
	boolStr := strconv.FormatBool(isItemOnTable)

	Response(w, boolStr, http.StatusOK, nil)
}

// ItemOutputComplete - [API] 입주민이 택배를 수령해 테이블에 물건이 없을 경우 호출
func ItemOutputComplete(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	if err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose); err != nil {
		// TODO - PLC 에러 처리
		log.Error(err)
		Response(w, nil, http.StatusInternalServerError, nil)
	}

	r.URL.Query().Get("itemId")
	itemIdStr := r.URL.Query().Get("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	// 트랜잭션
	tx, err := model.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	item, err := model.SelectItemById(itemId)
	if err != nil {
		log.Error(err)
		return
		// TODO - DB 에러 처리
	}

	// 트레이 업데이트
	trayUpdateRequest := model.TrayUpdateRequest{
		TrayOccupied: false,
		ItemId:       sql.NullInt64{Valid: false},
	}
	itemBottomSlot, err := model.SelectSlotByItemId(itemId)
	if err != nil {
		log.Error(err)
		return
		// TODO - DB 에러 처리
	}
	_, err = model.UpdateTrayEmpty(itemBottomSlot.TrayId.Int64, trayUpdateRequest, tx)
	if err != nil {
		log.Error(err)
		return
		// TODO - DB 에러 처리
	}

	slots, err := model.SelectSlotListByLaneAndItemId(itemId)
	if err != nil {
		log.Error(err)
		return
		// TODO - DB 에러 처리
	}

	// 물건이 차지하던 슬롯 초기화
	for idx := range slots {
		slot := &slots[idx]

		if slot.ItemId.Int64 == itemId {
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
		return
		// TODO - DB 에러 처리
	}

	// 택배 불출 시간 업데이트
	_, err = model.UpdateOutputTime(item.ItemId, tx)
	if err != nil {
		log.Error()
		return
		// TODO - DB 에러 처리
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	delete(requestList, itemId)

	if len(requestList) > 0 {
		// Request가 남아있는 경우- 택배가 나오고 있습니다 화면으로
		log.Info("[웹핸들러] 불출 요청이 남아 있어 택배 나오는 중 화면 출력")

		err := ChangeKioskView("/output/ongoing")
		if err != nil {
			log.Error(err)
			return
			// TODO - 에러 처리
		}

		err = RetrieveEmptyTrayFromTableAndUpdateDb()
		if err != nil {
			log.Error(err)
			return
			// TODO - 에러처리
		}

	} else {
		// Request가 남아있지 않은 경우
		log.Info("[웹핸들러] 불출 요청이 남아 있지 않아 감사합니다 화면 출력")

		err := ChangeKioskView("/output/thankyou")
		if err != nil {
			log.Error(err)
			// TODO - 에러 처리
			return
		}

		err = plc.DismissRobotAtTable()
		if err != nil {
			log.Error(err)
			return
			// TODO - PLC 에러 처리
		}
	}

	Response(w, nil, http.StatusOK, nil)
}
