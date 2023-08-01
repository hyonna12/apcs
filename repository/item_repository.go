package repository

import (
	"APCS/data/request"
	"APCS/data/response"
	"database/sql"
	"errors"
	log "github.com/sirupsen/logrus"
)

type ItemRepository struct {
	DB *sql.DB
}

func (i *ItemRepository) AssignDB(db *sql.DB) {
	i.DB = db
}

func (i *ItemRepository) SelectItemLocationList() (*[]response.ItemReadResponse, error) {
	var Resps []response.ItemReadResponse

	query := `SELECT i.item_id, s.lane, s.floor
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
			`

	rows, err := i.DB.Query(query)

	for rows.Next() {
		var Resp response.ItemReadResponse
		rows.Scan(&Resp.ItemId, &Resp.Lane, &Resp.Floor)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}

func (i *ItemRepository) SelectItemListByOwnerId(ownerId int) ([]response.ItemReadResponse, error) {
	var Resps []response.ItemReadResponse

	query := `SELECT i.item_id, i.item_name, i.item_height, s.lane, s.floor, s.tray_id
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
			WHERE i.owner_id = ? AND tray_id is not null
			`
	rows, err := i.DB.Query(query, ownerId)

	for rows.Next() {
		var Resp response.ItemReadResponse
		rows.Scan(&Resp.ItemId, &Resp.ItemName, &Resp.ItemHeight, &Resp.Lane, &Resp.Floor, &Resp.TrayId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return Resps, nil
	}
}

func (i *ItemRepository) SelectItemBySlot(lane, floor int) (*response.ItemReadResponse, error) {
	var Resp response.ItemReadResponse

	query := `SELECT i.item_id, i.item_name
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
			WHERE (s.lane = ? AND s.floor = ?)
			`
	err := i.DB.QueryRow(query, lane, floor).Scan(&Resp.ItemId, &Resp.ItemName)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("NOT FOUND")
		} else {
			return nil, err
		}
	} else {
		return &Resp, nil
	}
}

func (i *ItemRepository) InsertItem(req request.ItemCreateRequest) (sql.Result, error) {
	query := `INSERT INTO TN_CTR_ITEM(item_name, item_height, tracking_number, input_date, delivery_id, owner_id)
			VALUES(?, ?, ?, now(), ?, ?)
			`
	result, err := i.DB.Exec(query, req.ItemName, req.ItemHeight, req.TrackingNumber, req.DeliveryId, req.OwnerId)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, errors.New("NOT FOUND")
	}

	return result, nil
}

func (i *ItemRepository) UpdateOutputTime(ItemId int) (sql.Result, error) {
	query := `UPDATE TN_CTR_ITEM
			SET output_date = NOW()
			WHERE item_id = ?
			`
	result, err := i.DB.Exec(query, ItemId)

	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, errors.New("NOT FOUND")
	}

	return result, nil
}

func (i *ItemRepository) SelectItemIdByTrackingNum(tracking_number int) (response.ItemReadResponse, error) {
	var Resp response.ItemReadResponse
	log.Info(tracking_number)

	query := `SELECT item_id
			FROM TN_CTR_ITEM 
			WHERE tracking_number = ?
			`
	err := i.DB.QueryRow(query, tracking_number).Scan(&Resp.ItemId)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return Resp, errors.New("NOT FOUND")
		} else {
			return Resp, err
		}
	} else {
		return Resp, nil
	}
}
