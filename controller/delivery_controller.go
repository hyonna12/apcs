package controller

import (
	"APCS/data/response"
	"APCS/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func Response(w http.ResponseWriter, data interface{}, status int, err error) {
	var res response.CommonResponse

	if status == http.StatusOK {
		res.Data = data
		res.Status = status
	} else {
		res.Status = status
		res.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}
func DeliveryController(router *mux.Router) error {
	err := service.Service.InitService()

	if err != nil {
		return err
	}

	// GET 특정 id의 데이터 반환
	router.HandleFunc("/delivery/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])

		resp, err := service.Service.CheckDeliveryMatch(id)

		if err != nil {
			switch err.Error() {
			case "NOT FOUND":
				Response(w, nil, http.StatusNotFound, errors.New("해당 배달원이 없습니다."))
			default:
				Response(w, nil, http.StatusInternalServerError, err)
			}
			return
		}

		Response(w, resp, http.StatusOK, nil)

	}).Methods("GET")
}
