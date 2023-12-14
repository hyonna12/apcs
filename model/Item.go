package model

import (
	"apcs_refactored/customerror"
	"database/sql"
	"time"
)

type Item struct {
	ItemId         int64        `json:"item_id"`
	ItemHeight     int          `json:"item_height"`
	TrackingNumber int          `json:"tracking_number"`
	InputDate      sql.NullTime `json:"input_date"`
	OutputDate     sql.NullTime `json:"output_date"`
	DeliveryId     int64        `json:"delivery_id"`
	OwnerId        int64        `json:"owner_id"`
	CDatetime      sql.NullTime `json:"c_datetime"`
	UDatetime      sql.NullTime `json:"u_datetime"`
}

type ItemReadResponse struct {
	ItemId     int64 `json:"item_id"`
	ItemHeight int   `json:"item_height"`
	Lane       int   `json:"lane"`
	Floor      int   `json:"floor"`
	TrayId     int   `json:"tray_id"`
}

type ItemListResponse struct {
	ItemId          int64     `json:"item_id"`
	DeliveryCompany string    `json:"delivery_company"`
	TrackingNumber  int64     `json:"tracking_number"`
	InputDate       time.Time `json:"input_date"`
}

type ItemCreateRequest struct {
	ItemHeight     int   `json:"item_height"`
	TrackingNumber int   `json:"tracking_number"`
	DeliveryId     int64 `json:"delivery_id"`
	OwnerId        int64 `json:"owner_id"`
}

func SelectItemById(itemId int64) (Item, error) {

	query :=
		`
				SELECT 
					item_id,
					item_height,
					tracking_number,
					input_date,
					output_date,
					delivery_id,
					owner_id,
					c_datetime,
					u_datetime
				FROM TN_CTR_ITEM
				WHERE item_id = ?
			`

	var item Item
	row := DB.QueryRow(query, itemId)
	err := row.Scan(
		&item.ItemId,
		&item.ItemHeight,
		&item.TrackingNumber,
		&item.InputDate,
		&item.OutputDate,
		&item.DeliveryId,
		&item.OwnerId,
		&item.CDatetime,
		&item.UDatetime,
	)
	if err != nil {
		return Item{}, err
	}

	return item, nil
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

	rows, err := DB.Query(query)
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

	rows, err := DB.Query(query, ownerId)
	if err != nil {
		return nil, err
	}

	var itemReadResponses []ItemReadResponse

	for rows.Next() {
		var itemReadResponse ItemReadResponse
		err := rows.Scan(
			&itemReadResponse.ItemId,
			&itemReadResponse.ItemHeight,
			&itemReadResponse.Lane,
			&itemReadResponse.Floor,
			&itemReadResponse.TrayId,
		)
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

func SelectItemInfoByItemId(itemId int64) (ItemListResponse, error) {

	query := `
			SELECT
				i.item_id,
				d.delivery_company,
				i.tracking_number,
				i.input_date
			FROM TN_CTR_ITEM i
				JOIN TN_INF_OWNER	o ON i.owner_id = o.owner_id
				JOIN TN_INF_DELIVERY d ON i.delivery_id = d.delivery_id
			WHERE i.item_id = ?
			`

	var itemListResponse ItemListResponse
	row := DB.QueryRow(query, itemId)
	err := row.Scan(
		&itemListResponse.ItemId,
		&itemListResponse.DeliveryCompany,
		&itemListResponse.TrackingNumber,
		&itemListResponse.InputDate,
	)
	if err != nil {
		return ItemListResponse{}, err
	}

	return itemListResponse, nil
}

func SelectAddressByItemId(itemId int64) (string, error) {

	query := `
			SELECT
				o.address
			FROM TN_CTR_ITEM i
				JOIN TN_INF_OWNER	o ON i.owner_id = o.owner_id
			WHERE i.item_id = ?
			`

	var address string
	row := DB.QueryRow(query, itemId)
	err := row.Scan(
		&address,
	)
	if err != nil {
		return "", err
	}

	return address, nil
}

func SelectItemListByAddress(address string) ([]ItemListResponse, error) {

	query := `
			SELECT
				i.item_id,
				d.delivery_company,
				i.tracking_number,
				i.input_date
			FROM TN_CTR_ITEM i
				JOIN TN_INF_OWNER	o ON i.owner_id = o.owner_id
				JOIN TN_INF_DELIVERY d ON i.delivery_id = d.delivery_id
			WHERE 
			    o.address = ?
				AND
				i.output_date IS NULL
			`

	rows, err := DB.Query(query, address)
	if err != nil {
		return nil, err
	}

	var itemListResponses []ItemListResponse

	for rows.Next() {
		var itemListResponse ItemListResponse
		err := rows.Scan(
			&itemListResponse.ItemId,
			&itemListResponse.DeliveryCompany,
			&itemListResponse.TrackingNumber,
			&itemListResponse.InputDate,
		)
		if err != nil {
			return nil, err
		}
		itemListResponses = append(itemListResponses, itemListResponse)
	}

	if err != nil {
		return nil, err
	} else {
		return itemListResponses, nil
	}
}

func SelectItemBySlot(lane, floor int) (ItemReadResponse, error) {

	query :=
		`SELECT 
			i.item_id
		FROM TN_CTR_ITEM i
		JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
		WHERE (s.lane = ? AND s.floor = ?)
		`

	var itemReadResponse ItemReadResponse

	row := DB.QueryRow(query, lane, floor)
	err := row.Scan(
		&itemReadResponse.ItemId,
	)
	if err != nil {
		return ItemReadResponse{}, err
	}

	return itemReadResponse, nil
}

// InsertItem - DB에 물품 추가. 부여된 id 반환.
func InsertItem(itemCreateRequest ItemCreateRequest, tx *sql.Tx) (int64, error) {
	query := `INSERT INTO TN_CTR_ITEM(
                        item_height, 
                        tracking_number, 
                        input_date, 
                        delivery_id, 
                        owner_id)
			VALUES(?, ?, now(), ?, ?)
			`

	result, err := tx.Exec(query, itemCreateRequest.ItemHeight, itemCreateRequest.TrackingNumber, itemCreateRequest.DeliveryId, itemCreateRequest.OwnerId)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateOutputTime(itemId int64, tx *sql.Tx) (int64, error) {

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

func SelectItemIdByTrackingNum(trackingNumber string) (ItemListResponse, error) {
	query := `
			SELECT 
				i.item_id,
				d.delivery_company,
				i.tracking_number,
				i.input_date
			FROM TN_CTR_ITEM i
			JOIN TN_INF_OWNER	o ON i.owner_id = o.owner_id
			JOIN TN_INF_DELIVERY d ON i.delivery_id = d.delivery_id 
			WHERE tracking_number = ?
			AND i.output_date IS NULL
			`

	var itemReadResponse ItemListResponse

	row := DB.QueryRow(query, trackingNumber)
	err := row.Scan(&itemReadResponse.ItemId,
		&itemReadResponse.DeliveryCompany,
		&itemReadResponse.TrackingNumber,
		&itemReadResponse.InputDate)
	if err != nil {
		return ItemListResponse{}, err
	}

	return itemReadResponse, nil
}

func SelectItemExistsByAddress(address string) (bool, error) {
	query :=
		`SELECT EXISTS(
				SELECT 1
				FROM TN_CTR_ITEM i
					JOIN TN_INF_OWNER	o ON i.owner_id = o.owner_id
				WHERE 
				    o.address = ?
					AND
				    i.output_date IS NULL
				)
			`

	var exists bool
	row := DB.QueryRow(query, address)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, err
}

func SelectItemExistsByTrackingNum(trackingNumber string) (bool, error) {
	query :=
		`SELECT EXISTS(
				SELECT 1
				FROM TN_CTR_ITEM i
				JOIN TN_INF_OWNER	o 
				ON i.owner_id = o.owner_id
				WHERE i.tracking_number = ?
				AND i.output_date IS NULL
			)
			`

	var exists bool
	row := DB.QueryRow(query, trackingNumber)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, err
}

func SelectSortItemList() ([]ItemReadResponse, error) {
	query := `
			SELECT item_id, item_height 
			FROM TN_CTR_ITEM 
			WHERE output_date is null
			`

	var itemReadResponses []ItemReadResponse

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var itemReadResponse ItemReadResponse
		err := rows.Scan(&itemReadResponse.ItemId, &itemReadResponse.ItemHeight)
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

func SelectItemHeightByItemId(itemId int64) (ItemReadResponse, error) {

	query := `
			SELECT item_id, item_height 
			FROM TN_CTR_ITEM 
			WHERE item_id = ?
			`

	var itemReadResponses ItemReadResponse

	row := DB.QueryRow(query, itemId)
	err := row.Scan(&itemReadResponses.ItemId, &itemReadResponses.ItemHeight)
	if err != nil {
		return ItemReadResponse{}, err
	}

	return itemReadResponses, nil
}
