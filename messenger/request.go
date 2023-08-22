package messenger

const (
	KIOSK_REQUEST_TYPE_VIEW  = "view"
	KIOSK_REQUEST_TYPE_ALERT = "alert"
)

type KioskRequest struct {
	// Type 키오스크에 보내는 요청 타입
	//
	// messenger.KIOSK_REQUEST_TYPE_VIEW: 화면 전환
	// messenger.KIOSK_REQUEST_TYPE_ALERT: 경고창 띄우기
	Type string `json:"type"`
	// Data 타입에 따라 필요한 데이터
	//
	// Type == messenger.KIOSK_REQUEST_TYPE_VIEW 경우: view id
	// Type == messenger.KIOSK_REQUEST_TYPE_ALERT 경우: 메시지
	Data string `json:"data"`
}
