package handler

import (
	"apcs_refactored/model"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func CheckItemExists(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	log.Infof("[불출] 입주민 주소 입력: %v", address)

	exists, err := model.SelectItemExistsByAddress(address)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	if exists {
		_, err = fmt.Fprint(w, fmt.Sprintf("/output/item_list?address=%v", address))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Item Not Found", http.StatusNotFound)
	}
}
