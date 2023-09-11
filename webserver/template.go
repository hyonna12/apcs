package webserver

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var templ = func() *template.Template {
	t := template.New("")
	err := filepath.Walk("webserver/views/", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			log.Debug(path)
			_, err = t.ParseFiles(path)
			if err != nil {
				log.Error(err)
			}
		}
		return err
	})

	if err != nil {
		panic(err)
	}

	return t
}()

type Page struct {
	Title string
}

func Home(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "main", &Page{Title: "Home"})
}

/* Input_Item */
func RegistDelivery(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	templ.ExecuteTemplate(w, "input/regist_delivery", &Page{Title: "Home"})
}

func InputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	templ.ExecuteTemplate(w, "input/input_item", &Page{Title: "Home"})
}

func RegistOwner(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	templ.ExecuteTemplate(w, "input/regist_owner", &Page{Title: "Home"})
}

func InputItemError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	templ.ExecuteTemplate(w, "input/input_item_error", &Page{Title: "Home"})
}

func RegistOwnerError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	templ.ExecuteTemplate(w, "input/regist_owner_error", &Page{Title: "Home"})
}

func CompleteInputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	templ.ExecuteTemplate(w, "input/complete_input_item", &Page{Title: "Home"})
}

func CancelInputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	log.Infof("[웹핸들러] 수납 취소")

	// 빈 트레이 회수 및 DB 업데이트
	/* go func() {
		err := RetrieveEmptyTrayFromTableAndUpdateDb()
		if err != nil {
			log.Error(err)
			// TODO - 에러 처리
		}
	}() */

	templ.ExecuteTemplate(w, "input/cancel_input_item", &Page{Title: "Home"})
}

//
//func RegistAddressError(w http.ResponseWriter, r *http.Request) {
//	log.Debugf("URL: %v", r.URL)
//	if r.URL.Path != "/output/regist_address_error" {
//		http.Error(w, "Not found", http.StatusNotFound)
//		return
//	}
//
//	if r.Method != http.MethodGet {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//	}
//	templ.ExecuteTemplate(w, "output/regist_address_error", &Page{Title: "Home"})
//}

func ItemListError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/output/item_list_error" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "output/item_list_error", &Page{Title: "Home"})
}
