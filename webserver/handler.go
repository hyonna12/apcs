package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type CommonResponse struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
	Error  interface{} `json:"error"`
}

type InputInfoRequest struct {
	DeliveryId  string `json:"delivery_id"`
	Address     string `json:"address"`
	PhoneNum    string `json:"phone_num"`
	TrackingNum string `json:"tracking_num"`
}

type KioskRequest struct {
	RequestType KioskRequestType `json:"request_type"`
	Data        any              `json:"data"`
}

type ItemInfoData struct {
	ItemId          int64
	DeliveryCompany string
	TrackingNumber  int64
	InputDate       time.Time
}

type KioskRequestType string

const (
	kioskRequestTypeChangeView           = "changeView"
	kioskRequestTypeAlert                = "alert"
	KioskRequestCheckWebsocketConnection = "checkWebsocketConnection"
)

type RequestType string
type RequestStatus string

const (
	requestTypeInput  = "requestTypeInput"
	requestTypeOutput = "requestTypeOutput"
)

type request struct {
	itemId      int64
	requestType RequestType
}

var (
	requestList   map[int64]*request
	itemDimension plc.ItemDimension
	deliveryIdStr string
	ownerIdStr    string
	bestSlot      model.Slot
)

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

func ChangeKioskView(url string) error {
	KioskRequest := KioskRequest{
		RequestType: kioskRequestTypeChangeView,
		Data: struct {
			Url string `json:"url"`
		}{
			Url: url,
		},
	}
	request, err := json.Marshal(KioskRequest)
	if err != nil {
		log.Error(err)
		return err
	}
	broadcastToPrivate(request)

	return nil
}
func Alert(msg, url string) error {
	KioskRequest := KioskRequest{
		RequestType: kioskRequestTypeAlert,
		Data: struct {
			Msg string `json:"msg"`
			Url string `json:"url"`
		}{
			Msg: msg,
			Url: url,
		},
	}
	request, err := json.Marshal(KioskRequest)
	if err != nil {
		log.Error(err)
		return err
	}
	broadcastToPrivate(request)

	return nil
}

// RetrieveEmptyTrayFromTableAndUpdateDb
//
// 테이블의 빈 트레이를 회수.
// 빈 트레이를 격납할 위치를 선정하여 격납 후 DB 업데이트.
func RetrieveEmptyTrayFromTableAndUpdateDb() error {
	log.Info("[웹핸들러] 빈 트레이 회수")
	slots, err := model.SelectSlotListForEmptyTray()
	if err != nil {
		log.Error(err)
		return err
	}
	if len(slots) == 0 {
		err := errors.New("빈 슬롯 없음")
		// 1열이 아닌 다른 가까운 슬롯을 찾아서 넣은 후 불출 / 빈 슬롯 가져올 때 해당 슬롯에서 먼저 가져옴
		return err
	}
	// TODO - 빈 트레이 격납 위치 최적화
	slotForEmptyTray := slots[0]
	retrievedEmptyTrayId, err := plc.RetrieveEmptyTrayFromTable(slotForEmptyTray)
	if err != nil {
		return err
	}

	updateRequest := model.SlotUpdateRequest{
		SlotEnabled: true,
		SlotKeepCnt: slotForEmptyTray.SlotKeepCnt,
		TrayId:      sql.NullInt64{Int64: retrievedEmptyTrayId, Valid: true},
		ItemId:      sql.NullInt64{Valid: false},
		Lane:        slotForEmptyTray.Lane,
		Floor:       slotForEmptyTray.Floor,
	}

	tx, err := model.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	_, err = model.UpdateSlot(updateRequest, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// SenseTableForItem - [API] 테이블에 물품 여부 확인하기 위해 매 초마다 호출
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
