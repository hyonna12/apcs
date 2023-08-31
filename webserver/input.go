package webserver

import (
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/door"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	ownerId int64
	trayId  int64
)

type Data struct {
	Robot Robot `json:"robot"`
	Item  Item  `json:"item"`
}
type Robot struct {
	X string `json:"x"`
	Z string `json:"z"`
}
type Item struct {
	Height string `json:"height"`
	Weight string `json:"weight"`
}

func DeliveryCompanyList(w http.ResponseWriter, r *http.Request) {

	// 택배회사 리스트 조회
	deliveryList, err := model.SelectDeliveryCompanyList()
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	Response(w, deliveryList, http.StatusOK, nil)
}

func DeliveryInfoRequested(w http.ResponseWriter, r *http.Request) {
	inputInfoRequest = InputInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(&inputInfoRequest)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	if inputInfoRequest.Address == "" || inputInfoRequest.DeliveryId == "" {
		Response(w, nil, http.StatusBadRequest, errors.New("파라미터가 누락되었습니다"))
		return
	}

	ownerId, err = model.SelectOwnerIdByAddress(inputInfoRequest.Address)
	if ownerId == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("입력하신 주소가 존재하지 않습니다"))
		return
	}
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 테이블에 빈 트레이 감지
	emptyTray, err := plc.SenseTableForEmptyTray()
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}
	// 빈 트레이가 있을 경우
	if emptyTray {
		// tray_id 값 조회 **수정 - 트레이 큐알코드 스캔
		err := plc.StandbyRobotAtTable()
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}

		// 빈 트레이가 없을 경우
	} else {
		// 빈트레이를 가져올 슬롯 선택
		trayInfo, err := model.SelectEmptyTray()
		trayId = trayInfo.TrayId.Int64
		if trayId == 0 {
			Response(w, nil, http.StatusBadRequest, errors.New("빈 트레이가 존재하지 않습니다"))
			return
		}
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}

		err = plc.ServeEmptyTrayToTable(trayInfo)
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}

		slotupdate := model.SlotUpdateRequest{Lane: trayInfo.Lane, Floor: trayInfo.Floor}
		_, err = model.UpdateSlotToEmptyTray(slotupdate)
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}

		err = plc.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose)
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
	}

	Response(w, "/input/input_item", http.StatusOK, nil)
}

func ItemSubmitted(w http.ResponseWriter, r *http.Request) {
	err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationOpen)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 센싱하고 있다가 물품 감지
	/* for {
		IsItemOnTable, err := plc.SenseTableForItem() // 값 들어올때까지 대기
		time.Sleep(1 * time.Second)


		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		if IsItemOnTable {
			break
		}
	} */

	item, err := plc.SenseTableForItem() // 값 들어올때까지 대기
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}
	// **수정
	if !item {
		// 물품 크기, 무게, 송장번호 조회
		itemDimension, err = plc.SenseItemInfo()
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		itemDimension = plc.ItemDimension{Height: rand.Intn(6) + 1, Width: 5, Weigth: 8, TrackingNum: 1010} // **제거
		log.Printf("[제어서버] 아이템 크기/무게: %v", itemDimension)
	}

	// 물품의 크기, 무게가 기준 초과되면 입고 취소
	if itemDimension.Height > 10 {
		Response(w, nil, http.StatusBadRequest, errors.New("허용 높이 초과"))
		return
	}
	if itemDimension.Width > 10 {
		Response(w, nil, http.StatusBadRequest, errors.New("허용 너비 초과"))
		return
	}
	if itemDimension.Weigth > 10 {
		Response(w, nil, http.StatusBadRequest, errors.New("허용 무게 초과"))
		return
	}

	// 물품을 수납할 최적 슬롯 찾기 **수정
	data := Data{Robot: Robot{X: "10", Z: "1"}, Item: Item{Heigth: strconv.Itoa(itemDimension.Height), Weigth: strconv.Itoa(itemDimension.Weigth)}}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	resp, err := http.Post("http://localhost:8080/get/best_slot", "application/json", buff)

	if err != nil {
		// 에러나면 직접 수납슬롯 구하기
		fmt.Println(err)
		slotList, err := model.SelectAvailableSlotList(itemDimension.Height)
		if len(slotList) == 0 {
			Response(w, nil, http.StatusBadRequest, errors.New("수납가능한 슬롯이 존재하지 않습니다"))
			return
		}
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}

		sort.SliceStable(slotList, func(i, j int) bool {
			return slotList[i].TransportDistance < slotList[j].TransportDistance
		})
		bestSlot.Lane = slotList[0].Lane
		bestSlot.Floor = slotList[0].Floor
		fmt.Println("최적수납슬롯:", bestSlot)

	} else {
		defer resp.Body.Close()

		respData, err := io.ReadAll(resp.Body)
		json.Unmarshal(respData, &bestSlot)
		if err != nil {
			Response(w, nil, http.StatusBadRequest, errors.New("수납가능한 슬롯이 없습니다"))
			return
		}

		fmt.Println(bestSlot)
	}
	err = plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose)
	if err != nil {
		fmt.Println(err)
		Response(w, nil, http.StatusInternalServerError, err)
		return
	}
	Response(w, "/input/complete_input_item", http.StatusOK, nil)
}

func Input(w http.ResponseWriter, r *http.Request) {
	err := plc.InputItem(bestSlot)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 송장번호 ,물품높이, 택배기사, 수령인 정보 itemCreateRequest 에 넣어서 물품 db업데이트
	delivery_id, err := strconv.ParseInt(inputInfoRequest.DeliveryId, 10, 64)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}
	itemCreateRequest := model.ItemCreateRequest{ItemHeight: itemDimension.Height, TrackingNumber: itemDimension.TrackingNum, DeliveryId: delivery_id, OwnerId: ownerId}
	itemId, err := model.InsertItem(itemCreateRequest)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 슬롯, 트레이 db 업데이트
	// 트레이 아이디 추가
	trayUpdateRequest := model.TrayUpdateRequest{TrayOccupied: false, ItemId: itemId}
	_, err = model.UpdateTray(trayId, trayUpdateRequest)
	if err != nil {
		fmt.Println("1", err)
		//Response(w, nil, http.StatusInternalServerError, err)
	}
	_, err = model.UpdateStorageSlotKeepCnt(bestSlot.Lane, bestSlot.Floor, itemDimension.Height)
	if err != nil {
		fmt.Println("2")
		fmt.Println("밑에 빈 슬롯없음", err)
		//Response(w, nil, http.StatusInternalServerError, err)
	}
	slotUpdateRequest := model.SlotUpdateRequest{Lane: bestSlot.Lane, Floor: bestSlot.Floor, SlotEnabled: false, SlotKeepCnt: 0, TrayId: sql.NullInt64{Int64: trayId, Valid: true}, ItemId: sql.NullInt64{Int64: itemId, Valid: true}}
	_, err = model.UpdateStorageSlotList(itemDimension.Height, slotUpdateRequest)
	if err != nil {
		fmt.Println("3", err)
		//Response(w, nil, http.StatusInternalServerError, err)
	}
	_, err = model.UpdateSlot(slotUpdateRequest)
	if err != nil {
		fmt.Println("4", err)
		//Response(w, nil, http.StatusInternalServerError, err)
	}
	err = plc.SetUpDoor(door.DoorTypeBack, door.DoorOperationClose)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	Response(w, "OK", http.StatusOK, nil)
}

type StopRequest struct {
	Step string `json:"step"`
}

func StopInput(w http.ResponseWriter, r *http.Request) {
	stopRequest := StopRequest{}
	err := json.NewDecoder(r.Body).Decode(&stopRequest)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	if stopRequest.Step >= "2" {
		/* for {
			IsItemOnTable, err := plc.SenseTableForItem() // 값 들어올때까지 대기
			if err != nil {
				Response(w, nil, http.StatusInternalServerError, err)
			}
			if IsItemOnTable {
				break
			}
		} */
		// 센싱하고 있다가 물품 감지
		item, err := plc.SenseTableForItem() // 값 들어올때까지 대기
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		// **수정
		if !item {
			err := plc.SetUpDoor(door.DoorTypeFront, door.DoorOperationClose)
			if err != nil {
				Response(w, nil, http.StatusInternalServerError, err)
			}
		}
	}
	if stopRequest.Step >= "1" {
		Response(w, "OK", http.StatusOK, nil)
	}
}

func SenseItem(w http.ResponseWriter, r *http.Request) {
	fmt.Println("물품 감지")
	/* for {
		IsItemOnTable, err := plc.SenseTableForItem() // 값 들어올때까지 대기
		if err != nil {
			Response(w, nil, http.StatusInternalServerError, err)
		}
		if IsItemOnTable {
			break
		}
	} */
	item, err := plc.SenseTableForItem() // 값 들어올때까지 대기
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}
	// **수정

	if item == false {
		Response(w, "/input/input_item", http.StatusOK, nil)
	}
}

func Sort(w http.ResponseWriter, r *http.Request) {
	// 정리할 물품 선정 // **제거
	itemList, err := model.SelectSortItemList()
	if len(itemList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("정리가능한 물품이 존재하지 않습니다"))
		return
	}
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	item := itemList[rand.Intn(len(itemList)-1)]
	// 물품의 현재 슬롯
	currentSlot, err := model.SelectSlotByItemId(item.ItemId)
	fmt.Println("현재수납슬롯:", currentSlot)

	if len(itemList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("해당 물품이 보관되어 있지 않습니다"))
		return
	}
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// 이동할 슬롯 선정 // **제거
	slotList, err := model.SelectAvailableSlotList(item.ItemHeight)
	if len(slotList) == 0 {
		Response(w, nil, http.StatusBadRequest, errors.New("이동가능한 슬롯이 존재하지 않습니다"))
		return
	}
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	sort.SliceStable(slotList, func(i, j int) bool {
		return slotList[i].TransportDistance < slotList[j].TransportDistance
	})
	bestSlot.Lane = slotList[0].Lane
	bestSlot.Floor = slotList[0].Floor
	fmt.Println("최적수납슬롯:", bestSlot)

	// 트레이 이동
	err = plc.MoveTray(currentSlot, bestSlot)
	if err != nil {
		Response(w, nil, http.StatusInternalServerError, err)
	}

	// db 변경
	outputSlotUpdateRequest := model.SlotUpdateRequest{Lane: currentSlot.Lane, Floor: currentSlot.Floor, SlotEnabled: true}
	_, err = model.UpdateOutputSlotList(item.ItemHeight, outputSlotUpdateRequest)
	if err != nil {
		fmt.Println("1", err)
		//Response(w, nil, http.StatusInternalServerError, err)
	}
	_, err = model.UpdateOutputSlotKeepCnt(currentSlot.Lane, currentSlot.Floor)
	if err != nil {
		fmt.Println("2")
		fmt.Println("밑에 빈 슬롯없음", err)
	}
	_, err = model.UpdateSlotToEmptyTray(outputSlotUpdateRequest)
	if err != nil {
		fmt.Println("3", err)

		//Response(w, nil, http.StatusInternalServerError, err)
	}

	inputSlotUpdateRequest := model.SlotUpdateRequest{Lane: bestSlot.Lane, Floor: bestSlot.Floor, SlotEnabled: false, TrayId: currentSlot.TrayId, ItemId: currentSlot.ItemId}
	_, err = model.UpdateStorageSlotKeepCnt(bestSlot.Lane, bestSlot.Floor, item.ItemHeight)
	if err != nil {
		fmt.Println("4")

		fmt.Println("밑에 빈 슬롯없음", err)
	}
	_, err = model.UpdateStorageSlotList(item.ItemHeight, inputSlotUpdateRequest)
	if err != nil {
		fmt.Println("5", err)

		//Response(w, nil, http.StatusInternalServerError, err)
	}
	_, err = model.UpdateSlot(inputSlotUpdateRequest)
	if err != nil {
		fmt.Println("6", err)

		//Response(w, nil, http.StatusInternalServerError, err)
	}

	Response(w, "OK", http.StatusOK, nil)
}
