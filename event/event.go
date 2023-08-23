package event

import (
	"apcs_refactored/messenger"
	"encoding/json"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type eventName string

const (
	// DeliveryInfoRequested - 키오스크에서 택배기사가 입고 버튼을 누른 경우 발생
	//
	// Event.EventData: 없음
	DeliveryInfoRequested eventName = "DeliveryInfoRequested"

	// DeliveryIdSubmittedEvent - 키오스크에서 택배기사가 배송 번호를 입력하고 확인 버튼을 누른 경우 발생
	//
	// Event.EventData: 택배사 이름(string), 배송번호(int)
	DeliveryIdSubmittedEvent eventName = "DeliveryIdSubmittedEvent"

	// DeliveryIdValidatedEvent - 유효한 배송 번호일 경우 발생
	//
	// Event.EventData: 없음
	DeliveryIdValidatedEvent eventName = "DeliveryIdValidatedEvent"

	// DeliveryIdRejectedEvent - 유효하지 않은 배송 번호일 경우 발생
	//
	// Event.EventData: 없음
	DeliveryIdRejectedEvent eventName = "DeliveryIdRejectedEvent"

	// ItemSubmittedEvent - 입주민이 물건을 테이블에 올려놓고 확인 버튼을 누를 경우 발생
	//
	// Event.EventData: 없음
	ItemSubmittedEvent eventName = "ItemSubmittedEvent"

	// ItemInputRejectedEvent - 물건을 넣을 슬롯이 없을 경우 발생
	//
	// Event.EventData: 없음
	ItemInputRejectedEvent eventName = "ItemInputRejectedEvent"

	TableReadyForItemInputEvent eventName = "TableReadyForItemInputEvent"
	KioskFailEvent              eventName = "KioskFailEvent"
)

var (
	msgNode *messenger.Node
)

type Event struct {
	EventName eventName `json:"event_name"`
	EventData string    `json:"event_data"`
}

// StartEventServer
//
// 이벤트 디스패처 초기화
func StartEventServer(n *messenger.Node) {
	msgNode = n

	// 메신저 리스닝
	go msgNode.ListenMessages(
		func(m *messenger.Message) bool {
			log.Debugf("[%v] Received Message: %v", msgNode.Name, m)
			go handleMessage(m)
			return true
		},
	)
}

// DispatchEvent
//
// 발생한 이벤트를 처리할 핸들러를 호출
func DispatchEvent(event Event) error {
	log.Debugf("[Event_Dispatcher] Event occurred: %v, %v", event.EventName, event.EventData)

	// Reflection -> 이벤트 이름에 따라 핸들러 호출
	handler := Handler{}
	args := []reflect.Value{reflect.ValueOf(event)}
	go reflect.ValueOf(handler).MethodByName(string(event.EventName) + "Handler").Call(args)

	return nil
}

func handleMessage(message *messenger.Message) {
	// 메시지 파싱 및 이벤트 호출
	e := Event{}
	if err := json.Unmarshal([]byte(message.Data), &e); err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Trigger event by messageWrapper: %v", e.EventName)
	if err := DispatchEvent(e); err != nil {
		log.Error(err)
	}
}
