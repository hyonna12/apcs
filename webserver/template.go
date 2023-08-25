package webserver

import (
	"fmt"
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
			fmt.Println(path)
			_, err = t.ParseFiles(path)
			if err != nil {
				fmt.Println(err)
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
	if r.URL.Path != "/input/regist_delivery" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "input/regist_delivery", &Page{Title: "Home"})
}

func InputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/input/input_item" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "input/input_item", &Page{Title: "Home"})
}

func RegistOwner(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/input/regist_owner" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "input/regist_owner", &Page{Title: "Home"})
}

func InputItemError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/input/input_item_error" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "input/input_item_error", &Page{Title: "Home"})
}

func RegistOwnerError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/input/regist_owner_error" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "input/regist_owner_error", &Page{Title: "Home"})
}

func CompleteInputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/input/complete_input_item" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "input/complete_input_item", &Page{Title: "Home"})
}

func CancelInputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/input/cancel_input_item" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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

func CompleteOutputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/output/complete_output_item" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	templ.ExecuteTemplate(w, "output/complete_output_item", &Page{Title: "Home"})
}
