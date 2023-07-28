package repository

import (
	"APCS/data/request"
	"APCS/data/response"
	"database/sql"
	"errors"
	"fmt"
)

type TrayRepository struct {
	DB *sql.DB
}

func (t *TrayRepository) AssignDB(db *sql.DB) {
	t.DB = db
}

func (t *TrayRepository) SelectTrayList() (*[]response.TrayReadResponse, error) {
	var Resps []response.TrayReadResponse

	query := `
			SELECT t.tray_id, s.lane, s.floor, t.item_id 
			FROM TN_CTR_TRAY t
			JOIN TN_CTR_SLOT s
			ON t.tray_id = s.tray_id
			`
	rows, err := t.DB.Query(query)

	for rows.Next() {
		var Resp response.TrayReadResponse
		rows.Scan(&Resp.TrayId, &Resp.Lane, &Resp.Floor, &Resp.ItemId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}

func (t *TrayRepository) SelectEmptyTrayList() (*[]response.TrayReadResponse, error) {
	var Resps []response.TrayReadResponse

	query := `
			SELECT t.tray_id, s.lane, s.floor
			FROM TN_CTR_TRAY t
			JOIN TN_CTR_SLOT s
			ON t.tray_id = s.tray_id
			WHERE tray_occupied = 1
			`

	rows, err := t.DB.Query(query)

	for rows.Next() {
		var Resp response.TrayReadResponse
		rows.Scan(&Resp.TrayId, &Resp.Lane, &Resp.Floor)
		Resps = append(Resps, Resp)
	}
	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}

func (t *TrayRepository) UpdateTray(trayId int, req request.TrayUpdateRequest) (sql.Result, error) {

	query := `
			UPDATE TN_CTR_TRAY
			SET tray_occupied = ?, item_id = ?
			WHERE tray_id = ?
			`
	result, err := t.DB.Exec(query, req.TrayOccupied, req.ItemId, trayId)

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

func (t *TrayRepository) UpdateTrayEmpty(trayId int, req request.TrayUpdateRequest) (sql.Result, error) {

	query := `
			UPDATE TN_CTR_TRAY
			SET tray_occupied = ?, item_id = null
			WHERE tray_id = ?
			`
	result, err := t.DB.Exec(query, req.TrayOccupied, trayId)
	fmt.Println("트레이업데이트:", result)
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
