package handler

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

/* Output_Item */
func RegistAddress(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/output/regist_address" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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
	if r.URL.Path != "/output/item_list" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	err := templ.ExecuteTemplate(w, "output/item_list", &Page{Title: "Home"})
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func GetItemList(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/output/get_item_list" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

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
	if r.URL.Path != "/output/item_output_ongoing" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	log.Infof("[웹핸들러] 아이템을 불출할 슬롯:%v", slots)

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

	// 슬롯 정보로 PLC에 불출 요청
	for _, slot := range slots {
		go func(s model.Slot) {
			log.Debugf("[웹핸들러 -> PLC] 불출 요청. 슬롯 id=%v, 아이템 id=%v", s.SlotId, s.ItemId.Int64)
			err := plc.OutputItem(s)
			if err != nil {
				log.Error(err)
			}

			// 택배가 테이블에 올라가면 요청 목록에서 제거
			log.Debugf("[웹핸들러] 불출 완료. 슬롯 id=%v, 아이템 id=%v", s.SlotId, s.ItemId.Int64)
			// TODO - 화면 전환, 비밀번호 입력, 수령/반품 로직

			delete(requestList, s.ItemId.Int64)
		}(slot)
	}

	err = templ.ExecuteTemplate(w, "output/item_output_ongoing", &Page{Title: "Home"})
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
