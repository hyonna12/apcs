package model

import (
	"apcs_refactored/customerror"
	"database/sql"
	"log"
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
	ItemId          int64         `json:"item_id"`
	DeliveryCompany string        `json:"delivery_company"`
	Address         string        `json:"address"`
	PhoneNum        string        `json:"phone_num"`
	TrackingNumber  int64         `json:"tracking_number"`
	InputDate       time.Time     `json:"input_date"`
	OutputDate      sql.NullTime  `json:"output_date"`
	Lane            sql.NullInt64 `json:"lane"`
	Floor           sql.NullInt64 `json:"floor"`
}

type ItemCreateRequest struct {
	ItemHeight     int   `json:"item_height"`
	TrackingNumber int   `json:"tracking_number"`
	DeliveryId     int64 `json:"delivery_id"`
	OwnerId        int64 `json:"owner_id"`
}

type ItemOption struct {
	SearchOption string `json:"searchOption"`
	SearchText   string `json:"searchText"`
}

type ItemCntReq struct {
	StartDate string `json:"startDate"`
	LastDate  string `json:"lastDate"`
}

type ItemCntRes struct {
	InputCnt  int64 `json:"input_cnt"`
	OutputCnt int64 `json:"output_cnt"`
	StoreCnt  int64 `json:"store_cnt"`
}

type ItemCntDateRes struct {
	Date       time.Time `json:"date"`
	InputCnt   int       `json:"input_cnt"`
	OutputCnt  int       `json:"output_cnt"`
	StoreCnt   int       `json:"store_cnt"`
	InputDiff  int       `json:"input_diff"`
	OutputDiff int       `json:"output_diff"`
	StoreDiff  int       `json:"store_diff"`
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

func SelectStoreItemList(itemOption *ItemOption) ([]ItemListResponse, error) {
	searchOption := itemOption.SearchOption
	searchText := itemOption.SearchText

	query := `
			SELECT i.item_id, tracking_number, INPUT_DATE, d.delivery_company, o.address, o.phone_num, s.lane, s.floor
			FROM TN_CTR_ITEM i
			JOIN TN_INF_DELIVERY d
			ON i.delivery_id = d.delivery_id
			JOIN TN_INF_OWNER o
			ON i.owner_id = o.owner_id
			JOIN TN_CTR_SLOT s
			ON s.item_id = i.item_id
			WHERE output_date IS NULL
			AND s.tray_id IS NOT NULL
			`

	if searchText != "" {
		log.Println("search==========", searchText)
		if searchOption == "0" {
			query += "AND i.tracking_number LIKE '%" + searchText + "%'"
		} else if searchOption == "1" {
			query += "AND o.address LIKE '%" + searchText + "%'"
		}
	}

	var itemListResponses []ItemListResponse

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var itemListResponse ItemListResponse
		err := rows.Scan(&itemListResponse.ItemId, &itemListResponse.TrackingNumber, &itemListResponse.InputDate, &itemListResponse.DeliveryCompany, &itemListResponse.Address, &itemListResponse.PhoneNum, &itemListResponse.Lane, &itemListResponse.Floor)
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

func SelectItemList(itemOption *ItemOption) ([]ItemListResponse, error) {
	searchOption := itemOption.SearchOption
	searchText := itemOption.SearchText

	query := `
				SELECT i.item_id, i.tracking_number, i.INPUT_DATE, d.delivery_company, o.address, i.output_date, o.phone_num, s.lane, s.floor
				FROM TN_CTR_ITEM i
				JOIN TN_INF_DELIVERY d
				ON i.delivery_id = d.delivery_id
				JOIN TN_INF_OWNER o
				ON i.owner_id = o.owner_id
				LEFT OUTER JOIN TN_CTR_SLOT s
				ON i.item_id = s.item_id AND tray_id IS NOT null
			`
	if searchText != "" {
		log.Println("search==========", searchText)
		if searchOption == "0" {
			query += "WHERE i.tracking_number LIKE '%" + searchText + "%'"
		} else if searchOption == "1" {
			query += "WHERE o.address LIKE '%" + searchText + "%'"
		}
	}

	var itemListResponses []ItemListResponse

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var itemListResponse ItemListResponse
		err := rows.Scan(&itemListResponse.ItemId, &itemListResponse.TrackingNumber, &itemListResponse.InputDate, &itemListResponse.DeliveryCompany, &itemListResponse.Address, &itemListResponse.OutputDate, &itemListResponse.PhoneNum, &itemListResponse.Lane, &itemListResponse.Floor)
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

func SelectOutputItemList(itemOption *ItemOption) ([]ItemListResponse, error) {
	searchOption := itemOption.SearchOption
	searchText := itemOption.SearchText

	query := `
				SELECT i.item_id, i.tracking_number, i.INPUT_DATE, i.output_date, d.delivery_company, o.address, o.phone_num
				FROM TN_CTR_ITEM i
				JOIN TN_INF_DELIVERY d
				ON i.delivery_id = d.delivery_id
				JOIN TN_INF_OWNER o
				ON i.owner_id = o.owner_id
				WHERE output_date IS NOT NULL
			`

	if searchText != "" {
		log.Println("search==========", searchText)
		if searchOption == "0" {
			query += "AND i.tracking_number LIKE '%" + searchText + "%'"
		} else if searchOption == "1" {
			query += "AND o.address LIKE '%" + searchText + "%'"
		}
	}

	var itemListResponses []ItemListResponse

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var itemListResponse ItemListResponse
		err := rows.Scan(&itemListResponse.ItemId, &itemListResponse.TrackingNumber, &itemListResponse.InputDate, &itemListResponse.OutputDate, &itemListResponse.DeliveryCompany, &itemListResponse.Address, &itemListResponse.PhoneNum)
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

func SelectInputItemByUser(owner_id interface{}) ([]ItemListResponse, error) {
	query := `
				SELECT item_id, tracking_number, INPUT_DATE, d.delivery_company, o.address
				FROM TN_CTR_ITEM i
				JOIN TN_INF_DELIVERY d
				ON i.delivery_id = d.delivery_id
				JOIN TN_INF_OWNER o
				ON i.owner_id = o.owner_id
				WHERE i.owner_id = ?
			`

	var itemListResponses []ItemListResponse

	rows, err := DB.Query(query, owner_id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var itemListResponse ItemListResponse
		err := rows.Scan(&itemListResponse.ItemId, &itemListResponse.TrackingNumber, &itemListResponse.InputDate, &itemListResponse.DeliveryCompany, &itemListResponse.Address)
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

func SelectOutputItemByUser(owner_id interface{}) ([]ItemListResponse, error) {
	query := `
				SELECT item_id, tracking_number, INPUT_DATE, output_date, d.delivery_company, o.address
				FROM TN_CTR_ITEM i
				JOIN TN_INF_DELIVERY d
				ON i.delivery_id = d.delivery_id
				JOIN TN_INF_OWNER o
				ON i.owner_id = o.owner_id
				WHERE output_date IS NOT NULL
				AND i.owner_id = ?
			`

	var itemListResponses []ItemListResponse

	rows, err := DB.Query(query, owner_id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var itemListResponse ItemListResponse
		err := rows.Scan(&itemListResponse.ItemId, &itemListResponse.TrackingNumber, &itemListResponse.InputDate, &itemListResponse.OutputDate, &itemListResponse.DeliveryCompany, &itemListResponse.Address)
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

func SelectStoreItemByUser(owner_id interface{}) ([]ItemListResponse, error) {
	query := `
			SELECT i.item_id, tracking_number, INPUT_DATE, d.delivery_company, o.address, s.lane, s.floor
			FROM TN_CTR_ITEM i
			JOIN TN_INF_DELIVERY d
			ON i.delivery_id = d.delivery_id
			JOIN TN_INF_OWNER o
			ON i.owner_id = o.owner_id
			JOIN TN_CTR_SLOT s
			ON s.item_id = i.item_id
			WHERE output_date IS NULL
			AND s.tray_id IS NOT null
			AND i.owner_id = ?
			`

	var itemListResponses []ItemListResponse

	rows, err := DB.Query(query, owner_id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var itemListResponse ItemListResponse
		err := rows.Scan(&itemListResponse.ItemId, &itemListResponse.TrackingNumber, &itemListResponse.InputDate, &itemListResponse.DeliveryCompany, &itemListResponse.Address, &itemListResponse.Lane, &itemListResponse.Floor)
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
func SelectItemCnt(itemDate *ItemCntReq) (ItemCntRes, error) {
	query := `
			SELECT
				COALESCE(SUM(CASE WHEN INPUT_DATE IS NOT NULL THEN 1 ELSE 0 END), 0) AS input_cnt,
				COALESCE(SUM(CASE WHEN INPUT_DATE IS NOT NULL AND OUTPUT_DATE IS NULL THEN 1 ELSE 0 END), 0) AS store_cnt,
				COALESCE(SUM(CASE WHEN OUTPUT_DATE IS NOT NULL THEN 1 ELSE 0 END), 0) AS output_cnt
			FROM TN_CTR_ITEM
		`

	var itemCntRes ItemCntRes
	var err error
	var row *sql.Row

	if *itemDate == (ItemCntReq{}) {
		row = DB.QueryRow(query)

	} else {
		query += `WHERE (DATE(INPUT_DATE) BETWEEN ? AND ?)
		OR (DATE(OUTPUT_DATE) BETWEEN ? AND ?)`

		row = DB.QueryRow(query, itemDate.StartDate, itemDate.LastDate, itemDate.StartDate, itemDate.LastDate)
	}

	err = row.Scan(&itemCntRes.InputCnt, &itemCntRes.StoreCnt, &itemCntRes.OutputCnt)
	if err != nil {
		return ItemCntRes{}, err
	}
	log.Println("cnt", itemCntRes)

	return itemCntRes, err

}
func SelectItemCntByDate() ([]ItemCntDateRes, error) {

	query := `
			WITH RECURSIVE dates(date) AS (
				SELECT CURDATE()
				UNION ALL
				SELECT DATE_SUB(date, INTERVAL 1 DAY)
				FROM dates
				WHERE date > DATE_SUB(CURDATE(), INTERVAL 2 DAY)
			),
			daily_counts AS (
				SELECT
					d.date,
					(SELECT COUNT(*) FROM TN_CTR_ITEM WHERE DATE(INPUT_DATE) = d.date) AS input_cnt,
					(SELECT COUNT(*) FROM TN_CTR_ITEM WHERE DATE(OUTPUT_DATE) = d.date) AS output_cnt,
					(SELECT COUNT(*) FROM TN_CTR_ITEM WHERE DATE(INPUT_DATE) <= d.date 
						AND (OUTPUT_DATE IS NULL OR DATE(OUTPUT_DATE) > d.date)) AS store_cnt
				FROM dates d
			)
			SELECT 
				c.date,
				c.input_cnt,
				c.output_cnt,
				c.store_cnt,
				c.input_cnt - COALESCE(p.input_cnt, 0) AS input_diff,
				c.output_cnt - COALESCE(p.output_cnt, 0) AS output_diff,
				c.store_cnt - COALESCE(p.store_cnt, 0) AS store_diff
			FROM daily_counts c
			LEFT JOIN daily_counts p ON p.date = DATE_SUB(c.date, INTERVAL 1 DAY)
			ORDER BY c.date DESC
		`

	var itemCntDateResList []ItemCntDateRes
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item ItemCntDateRes
		err := rows.Scan(&item.Date, &item.InputCnt, &item.OutputCnt, &item.StoreCnt,
			&item.InputDiff, &item.OutputDiff, &item.StoreDiff)
		if err != nil {
			return nil, err
		}
		itemCntDateResList = append(itemCntDateResList, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Println("cnt", itemCntDateResList)
	return itemCntDateResList, nil

}
func SelectItemCntWeekly() ([]ItemCntDateRes, error) {

	query := `
			WITH RECURSIVE dates(date) AS (
				SELECT CURDATE()
				UNION ALL
				SELECT DATE_SUB(date, INTERVAL 1 DAY)
				FROM dates
				WHERE date > DATE_SUB(CURDATE(), INTERVAL 6 DAY)
			),
			daily_counts AS (
				SELECT
					d.date,
					(SELECT COUNT(*) FROM TN_CTR_ITEM WHERE DATE(INPUT_DATE) = d.date) AS input_cnt,
					(SELECT COUNT(*) FROM TN_CTR_ITEM WHERE DATE(OUTPUT_DATE) = d.date) AS output_cnt
				FROM dates d
			)
			SELECT 
				c.date,
				c.input_cnt,
				c.output_cnt
			FROM daily_counts c
			LEFT JOIN daily_counts p ON p.date = DATE_SUB(c.date, INTERVAL 1 DAY)
			ORDER BY c.date DESC
		`

	var itemCntDateResList []ItemCntDateRes
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item ItemCntDateRes
		err := rows.Scan(&item.Date, &item.InputCnt, &item.OutputCnt)
		if err != nil {
			return nil, err
		}
		itemCntDateResList = append(itemCntDateResList, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return itemCntDateResList, nil

}
func SelectItemCntMonthly() ([]ItemCntDateRes, error) {

	query := `
			WITH RECURSIVE months(date) AS (
				SELECT DATE_FORMAT(CURDATE(), '%Y-%m-01')
				UNION ALL
				SELECT DATE_SUB(date, INTERVAL 1 MONTH)
				FROM months
				WHERE date > DATE_SUB(DATE_FORMAT(CURDATE(), '%Y-%m-01'), INTERVAL 4 MONTH)
			),
			monthly_counts AS (
				SELECT
					m.date,
					(SELECT COUNT(*) FROM TN_CTR_ITEM WHERE DATE_FORMAT(INPUT_DATE, '%Y-%m-01') = m.date) AS input_cnt,
					(SELECT COUNT(*) FROM TN_CTR_ITEM WHERE DATE_FORMAT(OUTPUT_DATE, '%Y-%m-01') = m.date) AS output_cnt
				FROM months m
			)
			SELECT 
				DATE(c.date) AS date,
				c.input_cnt,
				c.output_cnt
			FROM monthly_counts c
			LEFT JOIN monthly_counts p ON p.date = DATE_SUB(c.date, INTERVAL 1 MONTH)
			ORDER BY c.date DESC
		`

	var itemCntDateResList []ItemCntDateRes
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item ItemCntDateRes
		err := rows.Scan(&item.Date, &item.InputCnt, &item.OutputCnt)
		if err != nil {
			return nil, err
		}
		itemCntDateResList = append(itemCntDateResList, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Println("cnt", itemCntDateResList)
	return itemCntDateResList, nil

}
