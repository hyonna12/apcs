package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	render(w, "main.html", nil)
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
	go func() {
		err := RetrieveEmptyTrayFromTableAndUpdateDb()
		if err != nil {
			log.Error(err)
			// TODO - 에러 처리
		}
	}()

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

// ItemOutputOngoing - [View] "택배가 나오는 중입니다" 화면 출력
func ItemOutputOngoing(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	// Get 요청인 경우 화면만 출력
	if r.Method == http.MethodGet {
		render(w, "output/item_output_ongoing.html", nil)
		return
	}

	// Post 요청인 경우
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	itemIdsStr := r.PostForm["item_id"]
	var itemIds []int64
	log.Infof("[웹핸들러] 아이템 불출 요청 접수. itemIds=%v", itemIdsStr)

	// 접수 요청된 택배 물품을 요청 리스트에 추가
	for _, idStr := range itemIdsStr {
		itemId, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Error(err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}

		req := &request{
			itemId:      itemId,
			requestType: requestTypeOutput,
			//requestStatus: requestStatusPending, // TODO - 삭제
		}

		requestList[itemId] = req
		itemIds = append(itemIds, itemId)
	}

	// item id로 물품이 보관된 슬롯 얻어오기
	slots, err := model.SelectSlotListByItemIds(itemIds)
	if err != nil {
		log.Error(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	var slotIds []int64
	for _, slot := range slots {
		slotIds = append(slotIds, slot.SlotId)
	}
	log.Infof("[웹핸들러] 아이템을 불출할 슬롯: %v", slotIds)

	if len(slotIds) == 0 {
		// 에러처리
		log.Error()
		return
	}

	// 테이블에 빈 트레이가 있는 경우 회수 요청
	emptyTrayExistsOnTable, err := plc.SenseTableForEmptyTray()
	if err != nil {
		// TODO - PLC 에러 처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if emptyTrayExistsOnTable {
		go func() {
			err := RetrieveEmptyTrayFromTableAndUpdateDb()
			if err != nil {
				log.Error(err)
				// TODO - 에러 처리
			}
		}()

		// Output 요청이 먼저 테이블을 점유하는 것을 방지
		time.Sleep(1 * time.Second)
	}

	render(w, "output/item_output_ongoing.html", nil)

	// 슬롯 정보로 PLC에 불출 요청
	for _, slot := range slots {
		go func(s model.Slot) {
			time.Sleep(1 * time.Second) // 키오스크 웹소켓 연결 대기
			log.Debugf("[웹핸들러 -> PLC] 불출 요청. 슬롯 id=%v, 아이템 id=%v", s.SlotId, s.ItemId.Int64)
			err := plc.OutputItem(s)
			if err != nil {
				log.Error(err)
				// TODO - 불출 실패한 물품은 요청에서 삭제 및 알림 처리
				return
			}

			// 택배가 테이블에 올라가면 요청 목록에서 제거
			log.Debugf("[웹핸들러] 불출 완료. slotId=%v, itemId=%v", s.SlotId, s.ItemId.Int64)
			// TODO - 수령/반품 화면 전환
			err = ChangeKioskView("/output/confirm?itemId=" + strconv.FormatInt(s.ItemId.Int64, 10))
			if err != nil {
				// TODO - 에러처리
				log.Error(err)
				return
			}

			// TODO - 수령/반납 화면으로 넘길 지 결정
			delete(requestList, s.ItemId.Int64)
		}(slot)
	}
}

// ItemOutputConfirm - [View] "택배를 확인해주세요" 화면 출력
func ItemOutputConfirm(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemIdStr := r.URL.Query().Get("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	// TODO - 에러 처리
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	var itemInfo model.ItemListResponse
	itemInfo, err = model.SelectItemInfoByItemId(itemId)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	itemInfoData := ItemInfoData{
		ItemId:          itemInfo.ItemId,
		DeliveryCompany: itemInfo.DeliveryCompany,
		TrackingNumber:  itemInfo.TrackingNumber,
		InputDate:       itemInfo.InputDate,
	}

	pageData := struct {
		Title        string
		ItemInfoData ItemInfoData
	}{
		"수령 확인창",
		itemInfoData,
	}

	render(w, "output/item_output_confirm.html", pageData)
}

// ItemOutputPasswordForm - [View] 비밀번호 입력 화면 출력
func ItemOutputPasswordForm(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	itemIdStr := r.URL.Query().Get("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		// TODO - 에러처리
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	// TODO - 수정
	address, err := model.SelectAddressByItemId(itemId)
	if err != nil {
		// TODO - DB 에러 발생 시 에러처리
		log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	pageData := struct {
		Title   string
		Address string
	}{
		"수령 확인창",
		address,
	}

	render(w, "output/item_output_password_form.html", pageData)
}

// ItemOutputAccept - [VIEW] "택배를 꺼내 주세요" 화면 출력
func ItemOutputAccept(w http.ResponseWriter, r *http.Request) {
	log.Debugf("URL: %v", r.URL)

	render(w, "output/item_output_accept.html", nil)
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
