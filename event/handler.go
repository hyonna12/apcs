package event

import (
	"apcs_refactored/messenger"
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"
)

type Handler struct{}

// DeliveryIdSubmittedEventHandler 송장 번호가 제출되었을 경우 처리하는 핸들러
func (h Handler) DeliveryIdSubmittedEventHandler(event Event) error {
	log.Info("배송 정보 확인 중")

	// 키오스크에 "배송 정보를 확인 중입니다." 표시
	if err := sendKioskMessage(
		messenger.KIOSK_REQUEST_TYPE_VIEW,
		"$view_validating_delivery_id",
	); err != nil {
		return err
	}

	// TODO - 배송 정보 확인 로직
	isValidDeliveryId := event.EventData != "" // TODO - temp, 송장 번호가 없으면 오류
	time.Sleep(500 * time.Millisecond)         // TODO - temp, 키오스크 화면 시연을 위해 딜레이
	log.Info("배송 정보가 유효합니다.")

	if isValidDeliveryId {
		// 배송 정보가 유효할 경우
		return DispatchEvent(Event{EventName: DeliveryIdValidatedEvent})
	} else {
		// 배송 정보가 유효하지 않을 경우
		if err := sendKioskMessage(
			messenger.KIOSK_REQUEST_TYPE_ALERT,
			"배송 정보가 유효하지 않습니다. 다시 입력해주세요.",
		); err != nil {
			return nil
		}
		if err := sendKioskMessage(
			messenger.KIOSK_REQUEST_TYPE_VIEW,
			"$view_delivery_id_form",
		); err != nil {
			return nil
		}
		return DispatchEvent(Event{EventName: DeliveryIdRejectedEvent})
	}
}

// DeliveryIdValidatedEventHandler 송장 번호가 유효하지 않을 경우 처리하는 핸들러
func (h Handler) DeliveryIdValidatedEventHandler(event Event) error {
	// TODO - 임시 슬롯
	if err := plc.ServeEmptyTrayToTable(model.Slot{}); err != nil {
		return err
	}

	return sendKioskMessage(messenger.KIOSK_REQUEST_TYPE_VIEW, "$view_item_submit_request")
}

// DeliveryIdRejectedEventHandler 제출된 송장 번호가 유효하지 않을 경우 처리하는 핸들러
func (h Handler) DeliveryIdRejectedEventHandler(event Event) error {
	return nil
}

// ItemSubmittedEventHandler 유저가 물품을 테이블에 올려놓고 확인 버튼을 눌렀을 경우 처리하는 핸들러
func (h Handler) ItemSubmittedEventHandler(event Event) error {
	if err := plc.SetUpDoor(plc.DoorTypeFront, plc.DoorOperationOpen); err != nil {
		return err
	}

	log.Info("[제어서버 -> PLC] 물품 계측 요청")
	itemDimension, err := plc.SenseItemInfo()
	if err != nil {
		return err
	}
	log.Info("[제어서버] 아이템 크기/무게: %v", itemDimension)

	// TODO - [DB] 슬롯 데이터 가져오기
	log.Info("[DB] 슬롯 데이터 가져오기")
	// TODO - 수납 가능 여부 확인
	log.Info("수납 가능합니다.")

	isInputable := true // TODO - temp
	if isInputable {
		log.Info("[제어서버 -> PLC] 물품 수납 요청")
		if err := plc.InputItem(model.Slot{}); err != nil {
			return err
		}

		if err := sendKioskMessage(
			messenger.KIOSK_REQUEST_TYPE_VIEW,
			"$view_item_input_complete",
		); err != nil {
			return err
		}
	} else {
		// TODO - 수납 불가능한 이유도 같이 전달
		if err := sendKioskMessage(
			messenger.KIOSK_REQUEST_TYPE_VIEW,
			"$view_item_input_rejected",
		); err != nil {
			return err
		}

		return DispatchEvent(Event{EventName: ItemInputRejectedEvent})
	}

	return nil
}

func (h Handler) ItemInputRejectedEventHandler(event Event) error {
	// TODO
	return nil
}

// sendKioskMessage 키오스크에 메시지 전달
// TODO - 키오스크 응답 처리
func sendKioskMessage(msgType, msgData string) error {
	// 키오스크 제어 메시지 생성
	request, err := json.Marshal(messenger.KioskRequest{
		Type: msgType,
		Data: msgData,
	})
	if err != nil {
		return err
	}

	// 메시지 생성
	msg := messenger.NewMessage(messenger.LeafKiosk, messenger.LeafEvent, string(request))
	if err != nil {
		return err
	}

	// 메시지 전파
	m, _ := json.Marshal(msg)
	return msgNode.SpreadMessage(m)
}
