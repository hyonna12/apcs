package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type CommonResponse struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
	Error  interface{} `json:"error"`
}

type InputInfoRequest struct {
	DeliveryId string `json:"delivery_id"`
	Address    string `json:"address"`
	PhoneNum   string `json:"phone_num"`
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
	requestList map[int64]*request
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

// RetrieveEmptyTrayFromTableAndUpdateDb
//
// 테이블의 빈 트레이를 회수.
// 빈 트레이를 격납할 위치를 선정하여 격납 후 DB 업데이트.
func RetrieveEmptyTrayFromTableAndUpdateDb() error {
	log.Info("[웹핸들러] 입고 취소 후 빈 트레이 회수")
	slots, err := model.SelectSlotListForEmptyTray()
	if err != nil {
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
