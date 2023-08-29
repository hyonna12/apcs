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
	r.HandleFunc("/input/regist_delivery", RegistDelivery)
	r.HandleFunc("/input/input_item", InputItem)
	r.HandleFunc("/input/input_item_error", InputItemError)
	r.HandleFunc("/input/regist_owner", RegistOwner)
	r.HandleFunc("/input/regist_owner_error", RegistOwnerError)
	r.HandleFunc("/input/complete_input_item", CompleteInputItem)
	r.HandleFunc("/input/cancel_input_item", CancelInputItem)

	r.HandleFunc("/input/get_delivery_list", DeliveryCompanyList)
	r.HandleFunc("/input/input_delivery_info", DeliveryInfoRequested).Methods(http.MethodPost)
	r.HandleFunc("/input/submit_item", ItemSubmitted).Methods(http.MethodPost)
	r.HandleFunc("/input/input", Input).Methods(http.MethodPost)
	r.HandleFunc("/input/stop_input", StopInput).Methods(http.MethodPost)
	r.HandleFunc("/input/senseItem", SenseItem).Methods(http.MethodPost)

	/* output */
	r.HandleFunc("/output/regist_address", RegistAddress)
	// 알림창으로 대체
	//r.HandleFunc("/output/regist_address_error", RegistAddressError)
	r.HandleFunc("/output/check_item_exists", CheckItemExists).Methods(http.MethodGet) // API
	r.HandleFunc("/output/item_list", ItemList).Methods(http.MethodGet)
	r.HandleFunc("/output/get_item_list", GetItemList).Methods(http.MethodGet) // API
	r.HandleFunc("/output/ongoing", ItemOutputOngoing)
	r.HandleFunc("/output/confirm", ItemOutputConfirm).Methods(http.MethodGet)
	r.HandleFunc("/output/password/submit", ItemOutputSubmitPassword).Methods(http.MethodGet)
	r.HandleFunc("/output/password/check", ItemOutputCheckPassword).Methods(http.MethodPost) // API
	r.HandleFunc("/output/accept", ItemOutputAccept).Methods(http.MethodGet)                 // API
	r.HandleFunc("/output/return", ItemOutputReturn).Methods(http.MethodPost)                // API
	r.HandleFunc("/output/cancel", ItemOutputCancel).Methods(http.MethodGet)                 // API

	// 알림창으로 대체
	//r.HandleFunc("/output/item_list_error", ItemListError)
	r.HandleFunc("/output/complete_output_item", CompleteOutputItem)

	/* sort */
	r.HandleFunc("/sort", Sort).Methods(http.MethodPost)

}
