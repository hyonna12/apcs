package trouble

// TroubleEventHandler 트러블 이벤트 처리 인터페이스
type TroubleEventHandler interface {
	HandleTroubleEvent(troubleType string, details map[string]interface{}) error
}

// TroubleType 트러블 타입 상수
const (
	NetworkDisconnected = "network_status"
	TroubleEvent        = "trouble_event"
)
