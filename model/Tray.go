package model

import (
	"apcs_refactored/customerror"
	"database/sql"
	"time"
)

type Tray struct {
	TrayId       int64
	TrayOccupied bool
	ItemId       *int64
	CDatetime    time.Time
	UDatetime    time.Time
}

type TrayReadResponse struct {
	TrayId int64
	Lane   int
	Floor  int
	ItemId *int64
}

type TrayUpdateRequest struct {
	TrayOccupied bool
	ItemId       sql.NullInt64
}

func SelectTrayList() ([]TrayReadResponse, error) {
	query := `
			SELECT 
			    t.tray_id, 
			    s.lane, 
			    s.floor, 
			    t.item_id 
			FROM TN_CTR_TRAY t
			JOIN TN_CTR_SLOT s
			ON t.tray_id = s.tray_id
			ORDER BY tray_id
			`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	//var item_id *int64

	var trayReadResponses []TrayReadResponse

	for rows.Next() {
		var trayReadResponse TrayReadResponse
		err := rows.Scan(&trayReadResponse.TrayId, &trayReadResponse.Lane, &trayReadResponse.Floor, &trayReadResponse.ItemId)
		if err != nil {
			return nil, err
		}

		trayReadResponses = append(trayReadResponses, trayReadResponse)
	}

	return trayReadResponses, nil
}

func SelectEmptyTrayList() ([]TrayReadResponse, error) {
	query := `
			SELECT t.tray_id, s.lane, s.floor
			FROM TN_CTR_TRAY t
			JOIN TN_CTR_SLOT s
			ON t.tray_id = s.tray_id
			WHERE tray_occupied = 0
			`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	var trayReadResponses []TrayReadResponse

	for rows.Next() {
		var trayReadResponse TrayReadResponse
		err := rows.Scan(&trayReadResponse.TrayId, &trayReadResponse.Lane, &trayReadResponse.Floor, &trayReadResponse.ItemId)
		if err != nil {
			return nil, err
		}
		trayReadResponses = append(trayReadResponses, trayReadResponse)
	}

	return trayReadResponses, nil
}

func UpdateTray(trayId int64, trayUpdateRequest TrayUpdateRequest, tx *sql.Tx) (int64, error) {

	query := `
			UPDATE TN_CTR_TRAY
			SET 
			    tray_occupied = ?, 
			    item_id = ?
			WHERE tray_id = ?
			`

	result, err := tx.Exec(query, trayUpdateRequest.TrayOccupied, trayUpdateRequest.ItemId, trayId)
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

func UpdateTrayEmpty(trayId int64, trayUpdateRequest TrayUpdateRequest, tx *sql.Tx) (int64, error) {

	query := `
			UPDATE TN_CTR_TRAY
			SET 
			    tray_occupied = ?, 
			    item_id = null
			WHERE tray_id = ?
			`

	result, err := DB.Exec(query, trayUpdateRequest.TrayOccupied, trayId)
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
