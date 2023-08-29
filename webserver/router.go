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
	r.HandleFunc("/input/input_delivery_info", DeliveryInfoRequested).Methods("POST")
	r.HandleFunc("/input/submit_item", ItemSubmitted).Methods("POST")
	r.HandleFunc("/input/input", Input).Methods("POST")
	r.HandleFunc("/input/stop_input", StopInput).Methods("POST")
	r.HandleFunc("/input/senseItem", SenseItem).Methods("POST")

	/* output */
	r.HandleFunc("/output/regist_address", RegistAddress)
	// 알림창으로 대체
	//r.HandleFunc("/output/regist_address_error", RegistAddressError)
	r.HandleFunc("/output/check_item_exists", CheckItemExists) // API
	r.HandleFunc("/output/item_list", ItemList)
	r.HandleFunc("/output/get_item_list", GetItemList) // API
	r.HandleFunc("/output/ongoing", ItemOutputOngoing)
	r.HandleFunc("/output/confirm", ItemOutputConfirm)
	r.HandleFunc("/output/password/submit", ItemOutputSubmitPassword)
	r.HandleFunc("/output/password/check", ItemOutputCheckPassword) // API
	r.HandleFunc("/output/accept", ItemOutputAccept)                // API
	r.HandleFunc("/output/return", ItemOutputReturn)                // API

	// 알림창으로 대체
	//r.HandleFunc("/output/item_list_error", ItemListError)
	r.HandleFunc("/output/complete_output_item", CompleteOutputItem)

	/* sort */
	r.HandleFunc("/sort", Sort).Methods("POST")

}
