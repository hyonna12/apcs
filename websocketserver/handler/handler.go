package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Handler(r *mux.Router) {
	r.HandleFunc("/", Home)
	/* r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsHub, w, r)
	}) */

	http.Handle("/", r)
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("/static/")))
	r.Handle("/static/", staticHandler)

	/* input */
	r.HandleFunc("/input/regist_delivery", RegistDelivery)
	r.HandleFunc("/input/input_item", InputItem)
	r.HandleFunc("/input/input_item_error", InputItemError)
	r.HandleFunc("/input/regist_owner", RegistOwner)
	r.HandleFunc("/input/regist_owner_error", RegistOwnerError)
	r.HandleFunc("/input/complete_input_item", CompleteInputItem)
	r.HandleFunc("/input/cancel_input_item", CancelInputItem)

	r.HandleFunc("/input/input_delivery_info", func(w http.ResponseWriter, r *http.Request) {
		//
		//deliveryCreateRequest := request.DeliveryCreateRequest{}
		//
		//err := json.NewDecoder(r.Body).Decode(&deliveryCreateRequest)
		//
		//if err != nil {
		//	Response(w, nil, http.StatusInternalServerError, err)
		//}
		//
		//if deliveryCreateRequest.DeliveryName == "" || deliveryCreateRequest.PhoneNum == "" || deliveryCreateRequest.DeliveryCompany == "" {
		//	Response(w, nil, http.StatusBadRequest, errors.New("파라미터가 누락되었습니다"))
		//	return
		//}
		//
		//err = service.Service.InsertDelivery(deliveryCreateRequest)
		//
		//if err != nil {
		//	Response(w, nil, http.StatusInternalServerError, err)
		//	return
		//}
		//
		//Response(w, "OK", http.StatusOK, nil)

	}).Methods("POST")

	/* output */
	r.HandleFunc("/output/regist_address", RegistAddress)
	r.HandleFunc("/output/regist_address_error", RegistAddressError)
	r.HandleFunc("/output/item_list", ItemList)
	r.HandleFunc("/output/item_list_error", ItemListError)
	r.HandleFunc("/output/complete_output_item", CompleteOutputItem)

}