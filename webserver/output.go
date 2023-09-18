package webserver

import (
	"apcs_refactored/config"
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/door"
	"apcs_refactored/plc/trayBuffer"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// CheckItemExists - [API] 동호수 입력 시 호출
func CheckItemExists(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("address") {
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
	} else {
		tracking_num := r.URL.Query().Get("tracking_num")
		log.Infof("[불출] 입주민 주소 입력: %v", tracking_num)

		exists, err := model.SelectItemExistsByTrackingNum(tracking_num)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		if exists {
			_, err = fmt.Fprint(w, fmt.Sprintf("/output/item_list?tracking_num=%v", tracking_num))
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Item Not Found", http.StatusNotFound)
		}
	}
}

// GetItemList - [API] 아이템 목록 반환
func GetItemList(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	if r.URL.Query().Has("address") {
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
	} else {
		tracking_num := r.URL.Query().Get("tracking_num")
		itemResponses, err := model.SelectItemIdByTrackingNum(tracking_num)
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		response := CommonResponse{
			Data:   itemResponses,
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
}

// ItemOutputOngoing - [View] "택배가 나오는 중입니다" 화면 출력
func ItemOutputOngoing(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	// Get 요청인 경우 화면만 출력
	if r.Method == http.MethodGet {
		render(w, "output/item_output_ongoing.html", nil)
		return
	}

	// Post 요청인 경우
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	itemIdsStr := r.PostForm["item_id"]
	var itemIds []int64
	log.Infof("[웹핸들러] 아이템 불출 요청 접수. itemIds=%v", itemIdsStr)

	// 접수 요청된 택배 물품을 요청 리스트에 추가
	for _, idStr := range itemIdsStr {
		itemId, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Error(err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}

		req := &request{
			itemId:      itemId,
			requestType: requestTypeOutput,
			//requestStatus: requestStatusPending, // TODO - 삭제
		}

		requestList[itemId] = req
		itemIds = append(itemIds, itemId)
	}

	// item id로 물품이 보관된 슬롯 얻어오기
	slots, err := model.SelectSlotListByItemIds(itemIds)
	if err != nil {
		log.Error(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	var slotIds []int64
	for _, slot := range slots {
		slotIds = append(slotIds, slot.SlotId)
	}
	log.Infof("[웹핸들러] 아이템을 불출할 슬롯: %v", slotIds)

	if len(slotIds) == 0 {
		// 에러처리
		log.Error()
		return
	}

	render(w, "output/item_output_ongoing.html", nil)

	// 트레이 버퍼 개수 조회 후 (20개-물품개수) 될 때까지 회수
	for trayBuffer.Buffer.Count() > (config.Config.Plc.TrayBuffer.Optimum - len(itemIds)) {
		err := RetrieveEmptyTrayFromTableAndUpdateDb()
		if err != nil {
			log.Error(err)
			// TODO - 에러 처리
		}

		trayBuffer.Buffer.Pop()
		num := trayBuffer.Buffer.Count()
		err = model.InsertBufferState(num)
		if err != nil {
			log.Error(err)
			// TODO - 에러 처리
		}

		trayId := trayBuffer.Buffer.Peek().(int64)
		plc.TrayIdOnTable.Int64 = trayId
	}

	// Output 요청이 먼저 테이블을 점유하는 것을 방지
	time.Sleep(1 * time.Second)

	// 슬롯 정보로 PLC에 불출 요청
	for _, slot := range slots {
		go func(s model.Slot) {
			time.Sleep(1 * time.Second) // 키오스크 웹소켓 연결 대기
			log.Debugf("[웹핸들러 -> PLC] 불출 요청. 슬롯 id=%v, 아이템 id=%v", s.SlotId, s.ItemId.Int64)

			err := plc.OutputItem(s)
			if err != nil {
				log.Error(err)
				// TODO - 불출 실패한 물품은 요청에서 삭제 및 알림 처리
				return
			}

			trayBuffer.Buffer.Push(s.TrayId.Int64)
			num := trayBuffer.Buffer.Count()
			err = model.InsertBufferState(num)
			if err != nil {
				log.Error(err)
				// TODO - 에러 처리
			}

			plc.TrayIdOnTable.Int64 = s.TrayId.Int64

			// 택배가 테이블에 올라가면 요청 목록에서 제거
			log.Debugf("[웹핸들러] 불출 완료. slotId=%v, itemId=%v", s.SlotId, s.ItemId.Int64)
			// TODO - 수령/반품 화면 전환
			err = ChangeKioskView("/output/confirm?itemId=" + strconv.FormatInt(s.ItemId.Int64, 10))
			if err != nil {
				// TODO - 에러처리
				log.Error(err)
				return
			}

			// TODO - 수령/반납 화면으로 넘길 지 결정
			delete(requestList, s.ItemId.Int64)
		}(slot)
	}

}

// ItemOutputConfirm - [View] "택배를 확인해주세요" 화면 출력
func ItemOutputConfirm(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemIdStr := r.URL.Query().Get("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	// TODO - 에러 처리
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	var itemInfo model.ItemListResponse
	itemInfo, err = model.SelectItemInfoByItemId(itemId)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	itemInfoData := ItemInfoData{
		ItemId:          itemInfo.ItemId,
		DeliveryCompany: itemInfo.DeliveryCompany,
		TrackingNumber:  itemInfo.TrackingNumber,
		InputDate:       itemInfo.InputDate,
	}

	pageData := struct {
		Title        string
		ItemInfoData ItemInfoData
	}{
		"수령 확인창",
		itemInfoData,
	}

	render(w, "output/item_output_confirm.html", pageData)
}

// ItemOutputPasswordForm - [View] 비밀번호 입력 화면 출력
func ItemOutputPasswordForm(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemIdStr := r.URL.Query().Get("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		// TODO - 에러처리
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	// TODO - 수정
	address, err := model.SelectAddressByItemId(itemId)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	pageData := struct {
		Title   string
		Address string
	}{
		"수령 확인창",
		address,
	}

	render(w, "output/item_output_password_form.html", pageData)
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

	adminPassword, err := model.SelectAdminPassword()
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// 마스터 pw 값으로 수정***
	if request.Password == password || request.Password == adminPassword {
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

		trayBuffer.Buffer.Pop()
		num := trayBuffer.Buffer.Count()
		err = model.InsertBufferState(num)
		if err != nil {
			log.Error(err)
			// TODO - 에러 처리
		}

		trayId := trayBuffer.Buffer.Peek().(int64)
		plc.TrayIdOnTable.Int64 = trayId
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

			trayBuffer.Buffer.Pop()
			num := trayBuffer.Buffer.Count()
			err = model.InsertBufferState(num)
			if err != nil {
				log.Error(err)
				// TODO - 에러 처리
			}

			trayId := trayBuffer.Buffer.Peek().(int64)
			plc.TrayIdOnTable.Int64 = trayId
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

	} else {
		// Request가 남아있지 않은 경우
		log.Info("[웹핸들러] 불출 요청이 남아 있지 않아 감사합니다 화면 출력")

		err := ChangeKioskView("/output/thankyou")
		if err != nil {
			log.Error(err)
			// TODO - 에러 처리
			return
		}
	}

	err = plc.DismissRobotAtTable()
	if err != nil {
		log.Error(err)
		return
		// TODO - PLC 에러 처리
	}

	Response(w, nil, http.StatusOK, nil)
}
