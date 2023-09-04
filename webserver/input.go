package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/door"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
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
		Response(w, nil, http.StatusInternalServerError, err)
	}

	Response(w, deliveryList, http.StatusOK, nil)
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

	if inputInfoRequest.Address == "" || inputInfoRequest.DeliveryId == "" {
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
		log.Error(err)
		Response(w, nil, http.StatusInternalServerError, err)
		return
	}

	// 테이블에 빈 트레이 감지
	emptyTray, err := plc.SenseTableForEmptyTray()
	if err != nil {
		log.Error(err)
		Response(w, nil, http.StatusInternalServerError, err)
		return
	}
	// 빈 트레이가 있을 경우
	if emptyTray {
		log.Infof("[웹 핸들러] 테이블에 빈 트레이가 있어 사용. trayId=%v", emptyTray)

		// tray_id 값 조회
		//trayId, err := plc.GetTrayIdOnTable()
		//if err != nil {
		//	if err.Error() == customerror.ErrNoEmptyTrayOnTable {
		//		// TODO - 에러처리
		//		log.Error(err)
		//	}
		//}

		err = plc.StandbyRobotAtTable()
		if err != nil {
			log.Error(err)
			Response(w, nil, http.StatusInternalServerError, err)
		}

		// 빈 트레이가 없을 경우
	} else {
		// 빈트레이를 가져올 슬롯 선택
		slotWithEmptyTray, err := model.SelectSlotWithEmptyTray()
		trayId := slotWithEmptyTray.TrayId.Int64
		log.Infof("[웹 핸들러] 빈 트레이를 가져올 slotId=%v, trayId=%v", slotWithEmptyTray.SlotId, trayId)
		if trayId == 0 {
			log.Info("[웹 핸들러] 빈 트레이가 존재하지 않음")
			Response(w, nil, http.StatusBadRequest, errors.New("빈 트레이가 존재하지 않습니다"))
			return
		}
		if err != nil {
			log.Error(err)
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}

		err = plc.ServeEmptyTrayToTable(slotWithEmptyTray)
		if err != nil {
			log.Error(err)
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}

		slotUpdateRequest := model.SlotUpdateRequest{
			Lane:  slotWithEmptyTray.Lane,
			Floor: slotWithEmptyTray.Floor,
		}
		_, err = model.UpdateSlotToEmptyTray(slotUpdateRequest)
		if err != nil {
			log.Error(err)
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}

		err = plc.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose)
		if err != nil {
			log.Error(err)
			Response(w, nil, http.StatusInternalServerError, err)
			return
		}
	}
	log.Infof("[웹 핸들러] OwnerId=%v", ownerId)
	ownerIdStr := strconv.FormatInt(ownerId, 10)
	redirectUrl := "/input/input_item?deliveryId=" + inputInfoRequest.DeliveryId + "&ownerId=" + ownerIdStr

	Response(w, redirectUrl, http.StatusOK, nil)
}

// ItemSubmitted
//
// [API] 택배기사가 물건을 테이블에 올려놓은 경우 호출
func ItemSubmitted(w http.ResponseWriter, r *http.Request) {
	err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	deliveryIdStr := r.URL.Query().Get("deliveryId")
	ownerIdStr := r.URL.Query().Get("ownerId")

	// 센싱하고 있다가 물품 감지
	/* for {
		IsItemOnTable, err := plc.SenseTableForItem() // 값 들어올때까지 대기
		time.Sleep(1 * time.Second)


		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		if IsItemOnTable {
			break
		}
	} */

	isItemOnTable, err := plc.SenseTableForItem() // 값 들어올때까지 대기
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// TODO - temp
	var itemDimension plc.ItemDimension
	// **수정
	if !isItemOnTable {
		// 물품 크기, 무게, 송장번호 조회
		itemDimension, err = plc.SenseItemInfo()
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		itemDimension = plc.ItemDimension{Height: rand.Intn(6) + 1, Width: 5, Weight: 8, TrackingNum: 1010} // **제거
		log.Printf("[제어서버] 아이템 크기/무게: %v", itemDimension)
	}

	// 물품의 크기, 무게가 기준 초과되면 입고 취소
	if itemDimension.Height > 10 {
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

	var bestSlot model.Slot

	if err != nil {
		// 에러나면 직접 수납슬롯 구하기
		log.Error(err)
		slotList, err := model.SelectAvailableSlotList(itemDimension.Height)
		if len(slotList) == 0 {
			Response(w, nil, http.StatusBadRequest, errors.New("수납가능한 슬롯이 존재하지 않습니다"))
			return
		}
		if err != nil {
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
		json.Unmarshal(respData, &bestSlot)
		if err != nil {
			Response(w, nil, http.StatusBadRequest, errors.New("수납가능한 슬롯이 없습니다"))
			return
		}

		log.Info("[웹핸들러] 수납 가능한 슬롯 없음")
	}

	Response(w, "/input/complete_input_item", http.StatusOK, nil)
	go inputItem(bestSlot, deliveryIdStr, ownerIdStr, itemDimension)
}

func inputItem(bestSlot model.Slot, deliveryIdStr string, ownerIdStr string, itemDimension plc.ItemDimension) {
	// 아이템 수납
	err := plc.InputItem(bestSlot)
	if err != nil {
		log.Error(err)
	}

	// 송장번호, 물품높이, 택배기사, 수령인 정보 itemCreateRequest 에 넣어서 물품 db업데이트
	deliveryId, err := strconv.ParseInt(deliveryIdStr, 10, 64)
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
	itemId, err := model.InsertItem(itemCreateRequest)
	if err != nil {
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
	_, err = model.UpdateTray(plc.GetTrayIdOnTable().Int64, trayUpdateRequest)
	if err != nil {
		// TODO - 에러처리
		log.Error(err)
		return
	}

	//_, err = model.UpdateStorageSlotKeepCnt(bestSlot.Lane, bestSlot.Floor, itemDimension.Height)
	//if err != nil {
	//	log.Errorf("[웹핸들러] 밑에 빈 슬롯없음. error=%v", err)
	//	return
	//}
	//slotUpdateRequest := model.SlotUpdateRequest{
	//	Lane:        bestSlot.Lane,
	//	Floor:       bestSlot.Floor,
	//	SlotEnabled: false,
	//	SlotKeepCnt: 0,
	//	TrayId:      plc.GetTrayIdOnTable(),
	//	ItemId:      sql.NullInt64{Int64: itemId, Valid: true},
	//}

	// 아이템이 수납된 lane 슬롯 업데이트
	slots, err := model.SelectSlotsInLane(bestSlot.Lane)
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
			slot.TrayId = plc.GetTrayIdOnTable()
			continue
		}

		// 물건이 차지하는 슬롯 갱신
		itemTopFloor := bestSlot.Floor - itemDimension.Height + 1
		if itemTopFloor <= slot.Floor && slot.Floor <= bestSlot.Floor {
			slot.SlotEnabled = false
			slot.SlotKeepCnt = 0
			slot.ItemId = sql.NullInt64{Int64: itemId, Valid: true}
			slot.TrayId = sql.NullInt64{Valid: false} // set null
			continue
		}
	}

	// TODO - 삭제
	//for idx := range slots {
	//	slot := &slots[idx]
	//
	//	if slot.ItemId.Int64 == itemId {
	//		slot.SlotEnabled = false
	//		slot.SlotKeepCnt = 0
	//		slot.ItemId = sql.NullInt64{Int64: itemId, Valid: true} // set null
	//		slot.TrayId = sql.NullInt64{Valid: false} // set null
	//	}
	//}

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

	_, err = model.UpdateSlots(slots)
	if err != nil {
		log.Error(err)
		// TODO - DB 에러 처리
	}
	//
	//_, err = model.UpdateStorageSlotList(itemDimension.Height, slotUpdateRequest)
	//if err != nil {
	//	// TODO - 에러처리
	//	log.Error(err)
	//	return
	//}
	//_, err = model.UpdateSlot(slotUpdateRequest)
	//if err != nil {
	//	// TODO - 에러처리
	//	log.Error(err)
	//	return
	//}

	plc.SetTrayIdOnTable(sql.NullInt64{Valid: false}) // set null
}

// Input
// TODO - 삭제
//
// [API] 물품 계측 후 수납 가능 시 호출 (입고가 완료되었습니다 화면에서 호출)
//func Input(w http.ResponseWriter, r *http.Request) {
//	err := plc.InputItem(bestSlot)
//	if err != nil {
//		Response(w, nil, http.StatusInternalServerError, err)
//	}
//
//	// 송장번호 ,물품높이, 택배기사, 수령인 정보 itemCreateRequest 에 넣어서 물품 db업데이트
//	delivery_id, err := strconv.ParseInt(inputInfoRequest.DeliveryId, 10, 64)
//	if err != nil {
//		Response(w, nil, http.StatusInternalServerError, err)
//	}
//	itemCreateRequest := model.ItemCreateRequest{ItemHeight: itemDimension.Height, TrackingNumber: itemDimension.TrackingNum, DeliveryId: delivery_id, OwnerId: ownerId}
//	itemId, err := model.InsertItem(itemCreateRequest)
//	if err != nil {
//		Response(w, nil, http.StatusInternalServerError, err)
//	}
//
//	// 슬롯, 트레이 db 업데이트
//	// 트레이 아이디 추가
//	trayUpdateRequest := model.TrayUpdateRequest{TrayOccupied: true, ItemId: itemId}
//	_, err = model.UpdateTray(trayId, trayUpdateRequest)
//	if err != nil {
//		log.Error(err)
//		//Response(w, nil, http.StatusInternalServerError, err)
//	}
//	_, err = model.UpdateStorageSlotKeepCnt(bestSlot.Lane, bestSlot.Floor, itemDimension.Height)
//	if err != nil {
//		log.Errorf("밑에 빈 슬롯없음. error=%v", err)
//		//Response(w, nil, http.StatusInternalServerError, err)
//	}
//	slotUpdateRequest := model.SlotUpdateRequest{Lane: bestSlot.Lane, Floor: bestSlot.Floor, SlotEnabled: false, SlotKeepCnt: 0, TrayId: sql.NullInt64{Int64: trayId, Valid: true}, ItemId: sql.NullInt64{Int64: itemId, Valid: true}}
//	_, err = model.UpdateStorageSlotList(itemDimension.Height, slotUpdateRequest)
//	if err != nil {
//		log.Error(err)
//		//Response(w, nil, http.StatusInternalServerError, err)
//	}
//	_, err = model.UpdateSlot(slotUpdateRequest)
//	if err != nil {
//		log.Error(err)
//		//Response(w, nil, http.StatusInternalServerError, err)
//	}
//	err = plc.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose)
//	if err != nil {
//		Response(w, nil, http.StatusInternalServerError, err)
//	}
//
//	Response(w, "OK", http.StatusOK, nil)
//}

type StopRequest struct {
	Step string `json:"step"`
}

func StopInput(w http.ResponseWriter, r *http.Request) {
	stopRequest := StopRequest{}
	err := json.NewDecoder(r.Body).Decode(&stopRequest)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	if stopRequest.Step >= "2" {
		/* for {
			IsItemOnTable, err := plc.SenseTableForItem() // 값 들어올때까지 대기
			if err != nil {
				Response(w, nil, http.StatusInternalServerError, err)
			}
			if IsItemOnTable {
				break
			}
		} */
		// 센싱하고 있다가 물품 감지
		item, err := plc.SenseTableForItem() // 값 들어올때까지 대기
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		// **수정
		if !item {
			err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose)
			if err != nil {
				Response(w, nil, http.StatusInternalServerError, err)
			}
		}
	}
	if stopRequest.Step >= "1" {
		Response(w, "OK", http.StatusOK, nil)
	}
}

func SenseItem(w http.ResponseWriter, r *http.Request) {
	log.Info("[웹핸들러] 물품 감지")
	/* for {
		IsItemOnTable, err := plc.SenseTableForItem() // 값 들어올때까지 대기
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		if IsItemOnTable {
			break
		}
	} */
	item, err := plc.SenseTableForItem() // 값 들어올때까지 대기
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}
	// **수정

	if item == false {
		Response(w, "/input/input_item", http.StatusOK, nil)
	}
}

func Sort(w http.ResponseWriter, r *http.Request) {
	// 정리할 물품 선정 // **제거
	itemList, err := model.SelectSortItemList()
	if len(itemList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("정리가능한 물품이 존재하지 않습니다"))
		return
	}
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	item := itemList[rand.Intn(len(itemList)-1)]
	// 물품의 현재 슬롯
	currentSlot, err := model.SelectSlotByItemId(item.ItemId)
	log.Infof("[웹핸들러] 현재수납슬롯: slotId=", currentSlot.SlotId)

	if len(itemList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("해당 물품이 보관되어 있지 않습니다"))
		return
	}
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 이동할 슬롯 선정 // **제거
	slotList, err := model.SelectAvailableSlotList(item.ItemHeight)
	if len(slotList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("이동가능한 슬롯이 존재하지 않습니다"))
		return
	}
	if err != nil {
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
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// db 변경
	outputSlotUpdateRequest := model.SlotUpdateRequest{Lane: currentSlot.Lane, Floor: currentSlot.Floor, SlotEnabled: true}
	_, err = model.UpdateOutputSlotList(item.ItemHeight, outputSlotUpdateRequest)
	if err != nil {
		log.Error(err)
		//Response(w, nil, http.StatusInternalServerError, err)
	}
	_, err = model.UpdateOutputSlotKeepCnt(currentSlot.Lane, currentSlot.Floor)
	if err != nil {
		log.Errorf("[웹핸들러] 밑에 빈 슬롯없음. error=%v", err)
	}
	_, err = model.UpdateSlotToEmptyTray(outputSlotUpdateRequest)
	if err != nil {
		log.Error(err)

		//Response(w, nil, http.StatusInternalServerError, err)
	}

	inputSlotUpdateRequest := model.SlotUpdateRequest{Lane: bestSlot.Lane, Floor: bestSlot.Floor, SlotEnabled: false, TrayId: currentSlot.TrayId, ItemId: currentSlot.ItemId}
	_, err = model.UpdateStorageSlotKeepCnt(bestSlot.Lane, bestSlot.Floor, item.ItemHeight)
	if err != nil {
		log.Errorf("[웹핸들러] 밑에 빈 슬롯없음. error=%v", err)
	}
	_, err = model.UpdateStorageSlotList(item.ItemHeight, inputSlotUpdateRequest)
	if err != nil {
		log.Error(err)

		//Response(w, nil, http.StatusInternalServerError, err)
	}
	_, err = model.UpdateSlot(inputSlotUpdateRequest)
	if err != nil {
		log.Error(err)

		//Response(w, nil, http.StatusInternalServerError, err)
	}

	Response(w, "OK", http.StatusOK, nil)
}
