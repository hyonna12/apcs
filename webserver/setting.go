package webserver

import (
	"apcs_refactored/model"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// CheckAddressPassword - [API] 비밀번호가 제출된 경우 호출
func CheckAddressPassword(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	request := &struct {
		Address  string `json:"address"`
		Password string `json:"password"`
	}{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// TODO - 비밀번호 해싱
	hash := sha512.New()
	hash.Write([]byte(request.Password))
	//hash.Write([]byte("salt"))
	hashPassword := hex.EncodeToString(hash.Sum(nil))

	password, err := model.SelectPasswordByAddress(request.Address)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	adminPassword, err := model.SelectAdminPassword()
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// 마스터 pw 값으로 수정***
	if hashPassword == password || hashPassword == adminPassword {
		// 비밀번호 일치
		Response(w, nil, http.StatusOK, nil)
	} else {
		Response(w, nil, http.StatusBadRequest, errors.New("잘못된 비밀번호입니다"))
	}
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	request := &struct {
		Address string `json:"address"`
	}{}

	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	owner, err := model.SelectOwnerDetailByAddress(request.Address)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	response := CommonResponse{
		Data:   owner,
		Status: http.StatusOK,
		Error:  nil,
	}

	Response(w, response, http.StatusOK, nil)

}

func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	request := &struct {
		Address  string `json:"address"`
		Password string `json:"password"`
	}{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// TODO - 비밀번호 해싱
	hash := sha512.New()
	hash.Write([]byte(request.Password))
	//hash.Write([]byte("salt"))
	hashPassword := hex.EncodeToString(hash.Sum(nil))

	password, err := model.SelectPasswordByAddress(request.Address)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	adminPassword, err := model.SelectAdminPassword()
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// 마스터 pw 값으로 수정***
	if hashPassword == password || hashPassword == adminPassword {
		// 비밀번호 일치
		Response(w, nil, http.StatusOK, nil)
	} else {
		Response(w, nil, http.StatusBadRequest, errors.New("잘못된 비밀번호입니다"))
	}
}
