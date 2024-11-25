package interfaces

// TroubleEventHandler 트러블 이벤트 처리 인터페이스
type TroubleEventHandler interface {
	HandleTroubleEvent(troubleType string, details map[string]interface{}) error
}

// KioskView 키오스크 화면 제어 인터페이스
type KioskView interface {
	ChangeView(view string) error
	SendEvent(event string) error
}
