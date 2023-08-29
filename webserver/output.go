package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/door"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

/* Output_Item */
func RegistAddress(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	err := templ.ExecuteTemplate(w, "output/regist_address", &Page{Title: "Home"})
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

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

func ItemList(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	err := templ.ExecuteTemplate(w, "output/item_list", &Page{Title: "Home"})
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

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

func ItemOutputOngoing(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	// Get 요청인 경우 화면만 출력
	if r.Method == http.MethodGet {
		err := templ.ExecuteTemplate(w, "output/item_output_ongoing", &Page{Title: "Home"})
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	itemIdsStr := r.PostForm["item_id"]
	var itemIds []int64
	log.Infof("[웹핸들러] 아이템 불출 요청 접수. Item ids=%v", itemIdsStr)

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
	slots, err := model.SelectSlotsByItemIds(itemIds)
	if err != nil {
		log.Error(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	var slotIds []int64
	for _, slot := range slots {
		slotIds = append(slotIds, slot.SlotId)
	}
	log.Infof("[웹핸들러] 아이템을 불출할 슬롯: %v", slotIds)

	// 테이블에 빈 트레이가 있는 경우 회수 요청
	emptyTrayExistsOnTable, err := plc.SenseTableForEmptyTray()
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	if emptyTrayExistsOnTable {
		// 빈 트레이를 넣을 슬롯 선정
		emptySlots, err := model.SelectEmptySlotList()
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		// TODO - 빈 트레이를 수납할 최적 슬롯 선정
		slotForEmptyTray := emptySlots[0]

		err = plc.RetrieveEmptyTrayFromTable(slotForEmptyTray)
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}

	err = templ.ExecuteTemplate(w, "output/item_output_ongoing", &Page{Title: "Home"})
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

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

			// 택배가 테이블에 올라가면 요청 목록에서 제거
			log.Debugf("[웹핸들러] 불출 완료. slotId=%v, itemId=%v", s.SlotId, s.ItemId.Int64)
			// TODO - 수령/반품 화면 전환
			err = ChangeKioskView("/output/confirm?itemId=" + strconv.FormatInt(s.ItemId.Int64, 10))
			if err != nil {
				// TODO - 에러처리
			}

			// TODO - 수령/반납 화면으로 넘길 지 결정
			delete(requestList, s.ItemId.Int64)
		}(slot)
	}
}

func ItemOutputConfirm(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemIdStr := r.URL.Query().Get("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	var itemInfo model.ItemListResponse
	itemInfo, err = model.SelectItemInfoByItemId(itemId)
	if err != nil {
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

	err = templ.ExecuteTemplate(w, "output/item_output_confirm", pageData)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func ItemOutputSubmitPassword(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemIdStr := r.URL.Query().Get("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	// TODO - 수정
	address, err := model.SelectAddressByItemId(itemId)
	if err != nil {
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

	err = templ.ExecuteTemplate(w, "output/item_output_submit_password", pageData)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

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
	password, err := model.SelectPasswordByItemId(1)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	if request.Password == password {
		Response(w, nil, http.StatusOK, nil)
	} else {
		Response(w, nil, http.StatusBadRequest, errors.New("잘못된 비밀번호입니다"))
	}
}

func ItemOutputAccept(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	err := templ.ExecuteTemplate(w, "output/item_output_accept", nil)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func ItemOutputReturn(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemId, err := strconv.ParseInt(r.URL.Query().Get("itemId"), 10, 64)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusBadRequest)
	}

	if err = plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose); err != nil {
		// TODO - 앞문 닫힘 불가 시 에러처리
	}
	if err != nil {
		// TODO - 앞문 고장 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	delete(requestList, itemId)

	// 택배 반환
	// TODO - 반환하는 김에 정리(최적 슬롯 알고리즘) - 꺼낸 슬롯의 원래 자리는 비어있다고 가정하고 선정
	slot, err := model.SelectSlotByItemId(itemId)
	if err != nil {
		// TODO - 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	go func() {
		if err = plc.InputItem(slot); err != nil {
			// TODO - 반환 불가 시 에러처리
			log.Error(err)
		}
	}()

	Response(w, nil, http.StatusOK, nil)

	if len(requestList) > 0 {
		// Request가 남아있는 경우- 택배가 나오고 있습니다 화면으로
		err := ChangeKioskView("/output/ongoing")
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		// Request가 남아있지 않은 경우 - 택배 찾기가 취소되었습니다 화면으로
		err := ChangeKioskView("/output/cancel")
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}

}

func ItemOutputCancel(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	err := templ.ExecuteTemplate(w, "output/item_output_canceled", nil)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
