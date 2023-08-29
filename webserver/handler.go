package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
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
	//requestStatusPending = "pending" // TODO - 삭제
	//requestStatusOngoing = "ongoing" // TODO - 삭제
)

type request struct {
	itemId      int64
	requestType RequestType
	//requestStatus RequestStatus // TODO - 삭제
}

var (
	inputInfoRequest InputInfoRequest
	itemDimension    plc.ItemDimension
	bestSlot         model.Slot
	requestList      map[int64]*request
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
