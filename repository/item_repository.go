package repository

import (
	"APCS/data/response"
	"database/sql"
	"errors"
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

func (i *ItemRepository) SelectItemListByOwnerId(ownerId int) (*[]response.ItemReadResponse, error) {
	var Resps []response.ItemReadResponse

	query := `SELECT i.item_id, i.item_name, s.lane, s.floor
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
			WHERE i.owner_id = ?
			`
	rows, err := i.DB.Query(query, ownerId)

	for rows.Next() {
		var Resp response.ItemReadResponse
		rows.Scan(&Resp.ItemId, &Resp.ItemName, &Resp.Lane, &Resp.Floor)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}

func (i *ItemRepository) SelectItemListBySlot(lane, floor int) (*[]response.ItemReadResponse, error) {
	var Resps []response.ItemReadResponse

	query := `SELECT i.item_id, i.item_name
			FROM TN_CTR_ITEM i
			JOIN TN_CTR_SLOT s
			ON i.item_id = s.item_id
			WHERE (s.lane = ? AND s.floor = ?)
			`
	rows, err := i.DB.Query(query, lane, floor)

	for rows.Next() {
		var Resp response.ItemReadResponse
		rows.Scan(&Resp.ItemId, &Resp.ItemName)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}

func (i *ItemRepository) UpdateOutputTime() (sql.Result, error) {
	query := `UPDATE TN_CTR_ITEM
			SET output_time = NOW()
			WHERE item_id
			`
	result, err := i.DB.Exec(query)

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
