package handler

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type CommonResponse struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
	Error  interface{} `json:"error"`
}

/* type Data struct {
	Url          string      `json:"url"`
	ResponseData interface{} `json:"responseData"`
} */

type InputInfoRequest struct {
	DeliveryCompany string `json:"delivery_company"`
	Address         string `json:"address"`
	PhoneNum        string `json:"phone_num"`
}

var inputInfoRequest InputInfoRequest
var itemDimension plc.ItemDimension
var bestSlot model.Slot

func Response(w http.ResponseWriter, data interface{}, status int, err error) {
	var res CommonResponse

	if status == http.StatusOK {
		res.Data = data
		res.Status = status
	} else {
		res.Status = status
		res.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

func DeliveryCompanyList(w http.ResponseWriter, r *http.Request) {
	// 택배회사 리스트 조회
	deliveryList, err := model.SelectDeliveryCompanyList()
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}
	fmt.Println(deliveryList)
	Response(w, deliveryList, http.StatusOK, nil)
}

func DeliveryInfoRequested(w http.ResponseWriter, r *http.Request) {
	inputInfoRequest = InputInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(&inputInfoRequest)
	fmt.Println(inputInfoRequest)

	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	if inputInfoRequest.Address == "" || inputInfoRequest.PhoneNum == "" || inputInfoRequest.DeliveryCompany == "" {
		Response(w, nil, http.StatusBadRequest, errors.New("파라미터가 누락되었습니다"))
		return
	}

	// address 가 존재하는 주소인지 확인하는 로직 필요! **수정 - id 값 받아오기

	// 테이블에 빈 트레이 감지
	emptyTray, _ := plc.SenseTableForEmptyTray()
	emptyTray = true // **제거
	// 빈 트레이가 있을 경우
	if emptyTray {
		// tray_id 값 조회 **수정
		plc.StandbyRobotAtTable()
		// 빈 트레이가 없을 경우
	} else {
		// 빈트레이를 가져올 슬롯 선택 - tray_id 값 조회 **수정
		emptyTray := model.Slot{}
		plc.ServeEmptyTrayToTable(emptyTray)
		// 트레이 db 정보 변경 **수정
		plc.SetUpDoor(plc.DoorTypeBack, plc.DoorOperationClose)
	}

	Response(w, "OK", http.StatusOK, nil)
}

func ItemSubmitted(w http.ResponseWriter, r *http.Request) {
	plc.SetUpDoor(plc.DoorTypeFront, plc.DoorOperationOpen)
	for {
		// 센싱하고 있다가 물품 감지
		item, _ := plc.SenseTableForItem() // 값 들어올때까지 대기
		item = true                        // **제거
		if item {
			// 물품 크기, 무게, 송장번호 조회
			itemDimension, _ = plc.SenseItemInfo()
			itemDimension = plc.ItemDimension{Height: 3, Width: 1, Length: 3, TrackingNum: 1010} // **제거
			log.Printf("[제어서버] 아이템 크기/무게: %v", itemDimension)
			break
		}
	}
	// 물품의 크기, 무게가 기준 초과되면 입고 취소
	if itemDimension.Height > 10 {
		Response(w, nil, http.StatusBadRequest, errors.New("허용 무게를 초과하였습니다"))
		// 중단 프로세스 **수정
		return
	}

	// 물품을 수납할 최적 슬롯 찾기
	bestSlot = model.Slot{SlotId: 1, Lane: 1, Floor: 1}

	// 수납할 수 있는 슬롯이 없을 때 입고 취소
	if bestSlot == (model.Slot{}) {
		Response(w, nil, http.StatusBadRequest, errors.New("수납 가능한 슬롯이 없습니다"))
		// 중단 프로세스 **수정
		return
	}

	plc.SetUpDoor(plc.DoorTypeFront, plc.DoorOperationClose)

	Response(w, "OK", http.StatusOK, nil)
}

func Input(w http.ResponseWriter, r *http.Request) {
	plc.InputItem(bestSlot)

	// ownerId 조회, deliveryId 조회 - 키오스크 입력 시 id 값이 넘어오도록 **수정
	deliveryId, _ := model.SelectDeliveryIdByCompany(inputInfoRequest.DeliveryCompany)
	ownerId, _ := model.SelectOwnerIdByAddress(inputInfoRequest.Address)
	// 송장번호 ,물품높이, 택배기사, 수령인 정보 itemCreateRequest 에 넣어서 물품 db업데이트
	itemCreateRequest := model.ItemCreateRequest{ItemHeight: itemDimension.Height, TrackingNumber: itemDimension.TrackingNum, DeliveryId: deliveryId, OwnerId: ownerId}
	model.InsertItem(itemCreateRequest)

	// 슬롯, 트레이 db 업데이트
	// 물품 id 조회
	item, _ := model.SelectItemIdByTrackingNum(itemDimension.TrackingNum)
	// 트레이 아이디 추가
	slotUpdateRequest := model.SlotUpdateRequest{SlotEnabled: false, SlotKeepCnt: 0, ItemId: item.ItemId}
	model.UpdateStorageSlotList(itemDimension.Height, slotUpdateRequest)
	model.UpdateStorageSlotKeepCnt(bestSlot.Lane, bestSlot.Floor, itemDimension.Height)

	plc.SetUpDoor(plc.DoorTypeBack, plc.DoorOperationClose)
}
