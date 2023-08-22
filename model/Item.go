package model

import (
	"apcs_refactored/customerror"
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"time"
)

type Item struct {
	ItemId         int64
	ItemName       string
	ItemHeight     int
	TrackingNumber int
	InputDate      time.Time
	OutputDate     time.Time
	DeliveryId     int64
	OwnerId        int64
	CDatetime      time.Time
	UDatetime      time.Time
}

type ItemReadResponse struct {
	ItemId     int64
	ItemName   string
	ItemHeight int
	Lane       int
	Floor      int
	TrayId     int
}

type ItemCreateRequest struct {
	ItemName       string
	ItemHeight     int
	TrackingNumber int
	DeliveryId     int64
	OwnerId        int64
}

func SelectItemLocationList() ([]ItemReadResponse, error) {

	query := `SELECT 
				i.item_id, 
				s.lane, 
				s.floor
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
				ON i.item_id = s.item_id
			`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var itemReadResponses []ItemReadResponse

	for rows.Next() {
		var itemReadResponse ItemReadResponse
		err := rows.Scan(&itemReadResponse.ItemId, &itemReadResponse.Lane, &itemReadResponse.Floor)
		if err != nil {
			return nil, err
		}
		itemReadResponses = append(itemReadResponses, itemReadResponse)
	}

	if err != nil {
		return nil, err
	} else {
		return itemReadResponses, nil
	}
}

func SelectItemListByOwnerId(ownerId int) ([]ItemReadResponse, error) {

	query := `SELECT 
				i.item_id, 
				i.item_name, 
				i.item_height, 
				s.lane, 
				s.floor, 
				s.tray_id
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
			WHERE 
			    i.owner_id = ? 
			  	AND tray_id is not null
			`

	rows, err := db.Query(query, ownerId)
	if err != nil {
		return nil, err
	}

	var itemReadResponses []ItemReadResponse

	for rows.Next() {
		var itemReadResponse ItemReadResponse
		err := rows.Scan(&itemReadResponse.ItemId, &itemReadResponse.Lane, &itemReadResponse.Floor)
		if err != nil {
			return nil, err
		}
		itemReadResponses = append(itemReadResponses, itemReadResponse)
	}

	if err != nil {
		return nil, err
	} else {
		return itemReadResponses, nil
	}
}

func SelectItemBySlot(lane, floor int) (ItemReadResponse, error) {

	query := `SELECT i.item_id, i.item_name
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
			WHERE (s.lane = ? AND s.floor = ?)
			`

	var itemReadResponse ItemReadResponse

	row := db.QueryRow(query, lane, floor)
	err := row.Scan(&itemReadResponse.ItemId, &itemReadResponse.ItemName)
	if err != nil {
		return ItemReadResponse{}, err
	}

	return itemReadResponse, nil
}

// InsertItem - DB에 물품 추가. 부여된 id 반환.
func InsertItem(itemCreateRequest ItemCreateRequest) (int64, error) {
	query := `INSERT INTO TN_CTR_ITEM(
                        item_name, 
                        item_height, 
                        tracking_number, 
                        input_date, 
                        delivery_id, 
                        owner_id)
			VALUES(?, ?, ?, now(), ?, ?)
			`

	result, err := db.Exec(query, itemCreateRequest.ItemName, itemCreateRequest.ItemHeight, itemCreateRequest.TrackingNumber, itemCreateRequest.DeliveryId, itemCreateRequest.OwnerId)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateOutputTime(itemId int) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Error(err)
		}
	}(tx)

	query := `
			UPDATE TN_CTR_ITEM
			SET output_date = NOW()
			WHERE item_id = ?
			`

	result, err := tx.Exec(query, itemId)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if affected == 0 {
		return 0, customerror.ErrNoRowsAffected
	}

	return affected, nil
}

func SelectItemIdByTrackingNum(trackingNumber int) (ItemReadResponse, error) {
	query := `SELECT item_id
			FROM TN_CTR_ITEM 
			WHERE tracking_number = ?
			`

	var itemReadResponse ItemReadResponse

	row := db.QueryRow(query, trackingNumber)
	err := row.Scan(&itemReadResponse.ItemId)
	if err != nil {
		return ItemReadResponse{}, err
	}

	return itemReadResponse, err
}
