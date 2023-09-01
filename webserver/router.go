package webserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Handler(r *mux.Router) {

	r.HandleFunc("/", Home)

	fileHandler := http.FileServer(http.Dir("./webserver/static"))
	stripHandler := http.StripPrefix("/static/", fileHandler)
	r.PathPrefix("/static/").Handler(stripHandler)

	// 웹소켓
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(wsHub, w, r)
	})

	/* input */
	r.HandleFunc("/input/regist_delivery", RegistDelivery).Methods(http.MethodGet)
	r.HandleFunc("/input/input_item", InputItem).Methods(http.MethodGet)
	r.HandleFunc("/input/input_item_error", InputItemError).Methods(http.MethodGet)
	r.HandleFunc("/input/regist_owner", RegistOwner).Methods(http.MethodGet)
	r.HandleFunc("/input/regist_owner_error", RegistOwnerError).Methods(http.MethodGet)
	r.HandleFunc("/input/complete_input_item", CompleteInputItem).Methods(http.MethodGet)
	r.HandleFunc("/input/cancel_input_item", CancelInputItem).Methods(http.MethodGet)

	r.HandleFunc("/input/get_delivery_list", DeliveryCompanyList).Methods(http.MethodGet)
	r.HandleFunc("/input/input_delivery_info", DeliveryInfoRequested).Methods(http.MethodPost)
	r.HandleFunc("/input/submit_item", ItemSubmitted).Methods(http.MethodPost)
	r.HandleFunc("/input/input", Input).Methods(http.MethodPost)
	r.HandleFunc("/input/stop_input", StopInput).Methods(http.MethodPost)
	r.HandleFunc("/input/senseItem", SenseItem).Methods(http.MethodPost)

	/* output */
	r.HandleFunc("/output/regist_address", RegistAddress)

	// [API] 동호수 입력 시 호출
	r.HandleFunc("/output/check_item_exists", CheckItemExists).Methods(http.MethodGet)

	// [View] 아이템 목록 화면 출력
	r.HandleFunc("/output/item_list", ItemList).Methods(http.MethodGet)

	// [API] 아이템 목록 반환
	r.HandleFunc("/output/get_item_list", GetItemList).Methods(http.MethodGet)

	// [View] "택배가 나오는 중입니다" 화면 출력
	r.HandleFunc("/output/ongoing", ItemOutputOngoing)

	// [View] "택배를 확인해주세요" 화면 출력
	r.HandleFunc("/output/confirm", ItemOutputConfirm).Methods(http.MethodGet)

	// [View] 비밀번호 입력 화면 출력
	r.HandleFunc("/output/password/form", ItemOutputPasswordForm).Methods(http.MethodGet)

	// [API] 비밀번호가 제출된 경우 호출
	r.HandleFunc("/output/password/check", ItemOutputCheckPassword).Methods(http.MethodPost)

	// [VIEW] "택배를 꺼내 주세요" 화면 출력
	r.HandleFunc("/output/accept", ItemOutputAccept).Methods(http.MethodGet)

	// [API] "택배를 확인해 주세요" 화면에서 "반납" 버튼을 누른 경우 호출
	// 비밀번호 입력 화면에서 "취소" 버튼을 누른 경우 호출
	// "택배를 꺼내 주세요" 화면에서 5초 경과 후 호출
	r.HandleFunc("/output/return", ItemOutputReturn).Methods(http.MethodPost)

	// [API] "택배를 꺼내주세요" 화면에서 매 초마다 호출
	r.HandleFunc("/output/sense_table_for_item", SenseTableForItem).Methods(http.MethodGet)

	// [VIEW] "택배 찾기가 취소되었습니다" 화면 출력
	// '/output/return' 호출 후 requestList에 요청이 남아 있지 않은 경우
	r.HandleFunc("/output/cancel", ItemOutputCancel).Methods(http.MethodGet)

	// [API] 입주민이 택배를 수령해 테이블에 물건이 없을 경우 호출
	r.HandleFunc("/output/complete", ItemOutputComplete).Methods(http.MethodPost)

	// [VIEW] "감사합니다" 화면 출력
	r.HandleFunc("/output/thankyou", ItemOutputThankyou).Methods(http.MethodGet)

	// TODO - temp - [API] 키오스크 물건 꺼내기 버튼 (시뮬레이션 용)
	r.HandleFunc("/output/takeout", ItemOutputTakeout).Methods(http.MethodPost)

	/* sort */
	r.HandleFunc("/sort", Sort).Methods(http.MethodPost)

}
