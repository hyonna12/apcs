package webserver

import (
	"apcs_refactored/config"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	// 템플릿 컬렉션
	tmpl *template.Template

	//tmplFS - HTML 템플릿 파일들을 바이너리 형태로 메모리에 올린 파일시스템 구현체(읽기 전용)
	//go:embed는 메모리로 퍼올릴 파일시스템 경로를 지정하는 컴파일러 지시어
	//go:embed views
	tmplFS embed.FS
)

func initTemplate() {
	// 새로운 템플릿 컬렉션 생성
	tmpl = template.New("")

	// views 디렉터리 아래 모든 .html 파일을 템플릿 컬렉션에 추가
	err := filepath.WalkDir("webserver/views/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 파일에 대해서만 템플릿 생성
		if d.IsDir() {
			return nil
		}

		// .html 파일에 대해서만 템플릿 생성
		if !strings.Contains(path, ".html") {
			return nil
		}

		// webserver/views 디렉터리 하위 상대경로 획득 (basepath는 프로젝트 루트 기준)
		relativePath, err := filepath.Rel("webserver/views", path)
		if err != nil {
			log.Panic(err)
			return err
		}

		// os가 windows인 경우 path 구분자를 백슬래시(\)에서 슬래시(/)로 변경
		relativePath = strings.Replace(relativePath, "\\", "/", -1)

		// embed.FS에서 파일 내용 byte 배열로 읽어오기
		data, err := tmplFS.ReadFile("views/" + relativePath)
		if err != nil {
			log.Panic(err)
			return err
		}

		// views 하위 상대경로를 이름으로 가지는 템플릿을 생성해서 컬렉션에 추가
		tmpl, err = tmpl.New(relativePath).Parse(string(data))
		if err != nil {
			log.Panic(err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

type Page struct {
	Title string
}

func render(w http.ResponseWriter, htmlFileName string, data any) {
	clonedTmpl, err := tmpl.Clone()
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	clonedTmpl, err = clonedTmpl.ParseFS(tmplFS, "views/"+htmlFileName)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = clonedTmpl.ExecuteTemplate(w, htmlFileName, data)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if config.Config.Kiosk.Ad == "off" {
		render(w, "main.html", nil)
	} else {
		render(w, "main_ad.html", nil)
	}
}

/* Input_Item */
func RegisterDelivery(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	render(w, "input/register_delivery.html", nil)
}

func InputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	render(w, "input/input_item.html", nil)
}

func RegisterOwner(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	render(w, "input/register_owner.html", nil)
}

func InputItemError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "input/input_item_error.html", nil)
}

func InputError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "input/input_error.html", nil)
}

func RegisterOwnerError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	render(w, "input/register_owner_error.html", nil)
}

func CompleteInputItem(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	render(w, "input/complete_input_item.html", nil)
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

	render(w, "input/cancel_input_item.html", nil)
}

//
//func RegistAddressError(w http.ResponseWriter, r *http.Request) {
//	log.Debugf("URL: %v", r.URL)
//	if r.URL.Path != "/output/register_address_error" {
//		http.Error(w, "Not found", http.StatusNotFound)
//		return
//	}
//
//	if r.Method != http.MethodGet {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//	}
// render(w, "output/register_address_error.html", nil)
//if err != nil {
//	http.Error(w, "InternalServerError", http.StatusInternalServerError)
//}
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
	render(w, "output/item_list_error.html", nil)
}

func RegistAddress(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/register_address.html", nil)
}

// ItemList - [View] 아이템 목록 화면 출력
func ItemList(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/item_list.html", nil)
}

// ItemOutputAccept - [VIEW] "택배를 꺼내 주세요" 화면 출력
func ItemOutputAccept(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/item_output_accept.html", nil)
}

// ItemOutputPasswordMismatch - [VIEW] "비밀번호가 일치하지 않습니다" 화면 출력
func ItemOutputPasswordMismatch(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/item_output_password_mismatch.html", nil)
}

// ItemOutputCancel
//
// [VIEW] "택배 찾기가 취소되었습니다" 화면 출력
// '/output/return' 호출 후 requestList에 요청이 남아 있지 않은 경우
func ItemOutputCancel(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/item_output_canceled.html", nil)
}

// ItemOutputThankyou - [VIEW] "감사합니다" 화면 출력
func ItemOutputThankyou(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/item_output_thankyou.html", nil)
}

func ItemError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/output/item_error" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	render(w, "output/item_error.html", nil)
}

func ReturnView(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)
	if r.URL.Path != "/output/item_return" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	render(w, "output/item_return.html", nil)
}

func Trouble(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "error/trouble.html", nil)
}

func OutputError(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/error.html", nil)
}
