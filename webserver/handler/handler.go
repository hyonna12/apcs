package handler

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type CommonResponse struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
	Error  interface{} `json:"error"`
}

type InputInfoRequest struct {
	DeliveryCompany string `json:"delivery_company"`
	Address         string `json:"address"`
	PhoneNum        string `json:"phone_num"`
}

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
	inputInfoRequest := InputInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(&inputInfoRequest)
	fmt.Println(inputInfoRequest)

	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	if inputInfoRequest.Address == "" || inputInfoRequest.PhoneNum == "" || inputInfoRequest.DeliveryCompany == "" {
		Response(w, nil, http.StatusBadRequest, errors.New("파라미터가 누락되었습니다"))
		return
	}

	// 테이블에 빈 트레이 감지
	emptyTray, _ := plc.SenseTableForEmptyTray()
	// 빈 트레이가 있을 경우
	if emptyTray {
		plc.StandbyRobotAtTable()
		// 빈 트레이가 없을 경우
	} else {
		// 트레이를 가져올 슬롯 선택
		// plc.ServeEmptyTrayToTable()
		// 트레이 db 정보 변경
		plc.SetUpDoor(plc.DoorTypeBack, plc.DoorOperationClose)
	}

	Response(w, "OK", http.StatusOK, nil)
}

/* func a() {
	plc.SetUpDoor(plc.DoorTypeFront, plc.DoorOperationOpen)
	item, _ := plc.SenseTableForItem()
	// 센싱하고 있다가 물품이 감지되면
	if item {
		plc.SetUpDoor(plc.DoorTypeFront, plc.DoorOperationOpen)
	}
}
*/
