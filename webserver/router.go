package webserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Handler(r *mux.Router) {

	r.HandleFunc("/", Home)

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("/static/")))
	r.Handle("/static/", staticHandler)

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

	/* output */
	r.HandleFunc("/output/regist_address", RegistAddress)
	// 알림창으로 대체
	//r.HandleFunc("/output/regist_address_error", RegistAddressError)
	r.HandleFunc("/output/check_item_exists", CheckItemExists)
	r.HandleFunc("/output/item_list", ItemList)
	r.HandleFunc("/output/get_item_list", GetItemList)
	r.HandleFunc("/output/item_output_ongoing", ItemOutputOngoing)
	r.HandleFunc("/output/item_output_confirm", ItemOutputConfirm)
	r.HandleFunc("/output/item_output_accept", ItemOutputAccept)

	// 알림창으로 대체
	//r.HandleFunc("/output/item_list_error", ItemListError)
	r.HandleFunc("/output/complete_output_item", CompleteOutputItem)

}
