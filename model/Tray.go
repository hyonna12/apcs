package model

import (
	"apcs_refactored/customerror"
	"context"
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
)

type Tray struct {
	TrayId       int64
	TrayOccupied bool
	ItemId       int64
	CDatetime    time.Time
	UDatetime    time.Time
}

type TrayReadResponse struct {
	TrayId int64
	Lane   int
	Floor  int
	ItemId int64
}

type TrayUpdateRequest struct {
	TrayOccupied bool
	ItemId       int64
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
			`
	rows, err := db.Query(query)
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

func SelectEmptyTrayList() ([]TrayReadResponse, error) {
	query := `
			SELECT t.tray_id, s.lane, s.floor
			FROM TN_CTR_TRAY t
			JOIN TN_CTR_SLOT s
			ON t.tray_id = s.tray_id
			WHERE tray_occupied = 1
			`

	rows, err := db.Query(query)
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

func UpdateTray(trayId int64, trayUpdateRequest TrayUpdateRequest) (int64, error) {
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
			UPDATE TN_CTR_TRAY
			SET 
			    tray_occupied = ?, 
			    item_id = ?
			WHERE tray_id = ?
			`

	result, err := db.Exec(query, trayUpdateRequest.TrayOccupied, trayUpdateRequest.ItemId, trayId)
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

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func UpdateTrayEmpty(trayId int, trayUpdateRequest TrayUpdateRequest) (int64, error) {
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
			UPDATE TN_CTR_TRAY
			SET 
			    tray_occupied = ?, 
			    item_id = null
			WHERE tray_id = ?
			`

	result, err := db.Exec(query, trayUpdateRequest.TrayOccupied, trayId)
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

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

// **제거
func SelectEmptyTray() (int64, error) {
	query := `
			SELECT MIN(tray_id)
			FROM TN_CTR_TRAY
			WHERE tray_occupied = 1
			`

	var trayId int64

	row := db.QueryRow(query)

	err := row.Scan(&trayId)
	if err != nil {
		log.Error(err)
		return trayId, err
	}

	return trayId, nil
}
