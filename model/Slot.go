package model

import (
	"apcs_refactored/customerror"
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

type SlotUpdateRequest struct {
	SlotEnabled bool
	SlotKeepCnt int
	TrayId      int64
	ItemId      int64
	Lane        int
	Floor       int
}

type Slot struct {
	SlotId            int64
	Lane              int
	Floor             int
	TransportDistance int
	SlotEnabled       bool
	SlotKeepCnt       int
	TrayId            int64
	ItemId            int64
	CheckDatetime     time.Time
	CDatetime         time.Time
	UDatetime         time.Time
}

func SelectSlotList() ([]Slot, error) {
	query := `
		SELECT slot_id, 
		       lane, 
		       floor, 
		       slot_keep_cnt, 
		       tray_id, 
		       item_id
			FROM TN_CTR_SLOT
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		err := rows.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.SlotKeepCnt, &slot.TrayId, &slot.ItemId)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

func SelectItemLocationByItemId(itemId int64) (Slot, error) {
	query := `
		SELECT 
			item_id, 
			slot_id, 
			lane, 
			floor 
		FROM TN_CTR_SLOT
		WHERE item_id = ?
		`

	var slot Slot

	row := db.QueryRow(query, itemId)
	err := row.Scan(&slot.ItemId, &slot.SlotId, &slot.Lane, &slot.Floor)
	if err != nil {
		log.Error(err)
		return Slot{}, err
	}

	return slot, nil
}

func SelectAvailableSlotList(itemHeight int) ([]Slot, error) {
	query := `
			SELECT 
				slot_id, 
				lane, 
				floor, 
				transport_distance, 
				tray_id 
			FROM TN_CTR_SLOT
			WHERE 
			    slot_enabled = 1 
			  	AND slot_keep_cnt >= ?
			`

	rows, err := db.Query(query, itemHeight)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		err := rows.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.SlotKeepCnt, &slot.TrayId, &slot.ItemId)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

func SelectSlotListWithoutItem() ([]Slot, error) {
	query := `
			SELECT slot_id, 
			       lane, 
			       floor, 
			       slot_keep_cnt, 
			       tray_id
			FROM TN_CTR_SLOT
			WHERE slot_enabled = 1
			`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		err := rows.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.SlotKeepCnt, &slot.TrayId, &slot.ItemId)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

func SelectEmptySlotList() ([]Slot, error) {
	query := `
			SELECT slot_id,
			       lane, 
			       floor,
			       transport_distance,
			       slot_enabled,
			       slot_keep_cnt,
			       tray_id,
			       item_id
			FROM TN_CTR_SLOT
			WHERE 
			    slot_enabled = 1 
				AND tray_id is null
			`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		err := rows.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.SlotKeepCnt, &slot.TrayId, &slot.ItemId)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

func SelectSlotListForEmptyTray() ([]Slot, error) {
	query := `
			SELECT 
			    s.slot_id, 
			    s.lane, 
			    s.floor, 
			    s.transport_distance, 
			    s.slot_keep_cnt, 
			    s.tray_id
			FROM TN_CTR_SLOT s
			JOIN TN_CTR_TRAY t
				ON s.tray_id = t.tray_id
			WHERE t.tray_occupied = 1
			`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		err := rows.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.SlotKeepCnt, &slot.TrayId, &slot.ItemId)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

func UpdateSlot(request SlotUpdateRequest) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	query := `
			UPDATE TN_CTR_SLOT
			SET 
			    slot_enabled = ?, 
			    slot_keep_cnt = ?, 
			    tray_id = ?, 
			    item_id = ?
			WHERE 
				lane = ? 
				AND floor = ?
			`

	result, err := tx.Exec(query, request.SlotEnabled, request.SlotKeepCnt, request.TrayId, request.ItemId, request.Lane, request.Floor)
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

//func UpdateStorageSlotList(itemHeight int, req SlotUpdateRequest) (int64, error) {
//	tx, err := db.BeginTx(context.Background(), nil)
//	if err != nil {
//		return 0, err
//	}
//
//	var minStorageSlot = req.Floor - itemHeight + 1
//	query := `
//			UPDATE TN_CTR_SLOT
//			SET slot_enabled = ?, slot_keep_cnt = ?, item_id = ?
//			WHERE (lane = ?) AND (floor >= ? AND floor <= ?)
//			`
//
//	result, err := tx.Exec(query, req.SlotEnabled, req.SlotKeepCnt, req.ItemId, req.Lane, minStorageSlot, req.Floor)
//	if err != nil {
//		return 0, err
//	}
//
//	affected, err := result.RowsAffected()
//	if err != nil {
//		return 0, err
//	}
//
//	if affected == 0 {
//		return 0, customerror.ErrNoRowsAffected
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		return 0, err
//	}
//
//	return affected, nil
//}

//
//func UpdateOutputSlotList(tx *sql.Tx, itemHeight int, req request.SlotUpdateRequest) (sql.Result, error) {
//	var minStorageSlot = (req.Floor - itemHeight + 1)
//	query := `
//			UPDATE TN_CTR_SLOT
//			SET slot_enabled = ?, slot_keep_cnt = floor, item_id = null
//			WHERE (lane = ?) AND (floor >= ? AND floor <= ?)
//			`
//
//	result, err := tx.Exec(query, req.SlotEnabled, req.Lane, minStorageSlot, req.Floor)
//	//result, err := s.DB.Exec(query, req.SlotEnabled, req.Lane, minStorageSlot, req.Floor)
//	if err != nil {
//		return nil, err
//	}
//
//	affected, err := result.RowsAffected()
//
//	if err != nil {
//		return nil, err
//	}
//
//	if affected == 0 {
//		return nil, errors.New("NOT FOUND")
//	}
//
//	return result, nil
//}
//
//func SelectStorageSlotListWithTray(itemHeight, lane, floor int) ([]response.SlotReadResponse, error) {
//	var minStorageSlot = (floor - itemHeight + 1)
//	var resps []response.SlotReadResponse
//
//	query := `
//			SELECT slot_id, lane, floor, tray_id
//			FROM TN_CTR_SLOT
//			WHERE (lane = ?) AND (floor >= ? AND floor <= ?) AND (tray_id IS NOT NULL)
//			`
//
//	rows, err := db.Query(query, lane, minStorageSlot, floor)
//
//	for rows.Next() {
//		var resp response.SlotReadResponse
//		rows.Scan(&resp.SlotId, &resp.Lane, &resp.Floor, &resp.TrayId)
//		resps = append(resps, resp)
//	}
//
//	if err != nil {
//		return nil, err
//	} else {
//		return resps, nil
//	}
//}
//func SelectSlotInfoByLocation(lane, floor int) (response.SlotReadResponse, error) {
//
//	var resp response.SlotReadResponse
//
//	query := `
//			SELECT slot_id, lane, floor, transport_distance, slot_enabled, slot_keep_cnt, tray_id, item_id
//			FROM TN_CTR_SLOT
//			WHERE (lane = ?) AND (floor  =?)
//			`
//	err := db.QueryRow(query, lane, floor).Scan(&resp.SlotId, &resp.Lane, &resp.Floor, &resp.TransportDistance, &resp.SlotEnabled, &resp.SlotKeepCnt, &resp.TrayId, &resp.ItemId)
//
//	if err != nil {
//		if err.Error() == "sql: no rows in result set" {
//			return resp, errors.New("NOT FOUND")
//		} else {
//			return resp, err
//		}
//	} else {
//		return resp, nil
//	}
//}
//
//func UpdateStorageSlotKeepCnt(tx *sql.Tx, lane, floor, itemHeight int) (sql.Result, error) {
//
//	query := `
//			UPDATE TN_CTR_SLOT s
//			SET s.slot_keep_cnt = (s.slot_keep_cnt - IF((s.floor=s.slot_keep_cnt), ?, ?))
//			WHERE (s.floor > ? AND s.floor <=
//				IFNULL(
//						(
//							SELECT * FROM (
//								SELECT MIN(floor) - 1
//								FROM TN_CTR_SLOT
//								WHERE (lane = ?) AND (floor > ? AND slot_keep_cnt = 0)
//							) a
//						),
//						(
//							SELECT * FROM (
//								SELECT MAX(floor)
//								FROM TN_CTR_SLOT
//								WHERE (lane = ? AND floor > ?)
//							) b
//						)
//					)
//				)
//			AND s.lane = ?
//			`
//
//	result, err := tx.Exec(query, floor, itemHeight, floor, lane, floor, lane, floor, lane)
//	//result, err := s.DB.Exec(query, floor, itemHeight, floor, lane, floor, lane, floor, lane)
//	if err != nil {
//		return nil, err
//	}
//
//	affected, err := result.RowsAffected()
//
//	if err != nil {
//		return nil, err
//	}
//
//	if affected == 0 {
//		return nil, errors.New("NOT FOUND")
//	}
//
//	return result, nil
//}
//
//func UpdateOutputSlotKeepCnt(tx *sql.Tx, lane, floor int) (sql.Result, error) {
//
//	query := `
//			UPDATE TN_CTR_SLOT s
//			SET s.slot_keep_cnt = (s.slot_keep_cnt +
//				IFNULL(
//						(
//							SELECT * FROM (
//								SELECT ? - MAX(FLOOR)
//								FROM TN_CTR_SLOT
//								WHERE (lane = ?) AND (FLOOR < ? AND slot_keep_cnt = 0)
//							) i
//						),
//							?
//						)
//					)
//			WHERE (s.floor > ? AND s.floor <=
//				IFNULL(
//						(
//							SELECT * FROM (
//								SELECT MIN(floor) - 1
//								FROM TN_CTR_SLOT
//								WHERE (lane = ?) AND (floor > ? AND slot_keep_cnt = 0)
//								) a
//							),
//							(
//								SELECT * FROM (
//									SELECT MAX(floor)
//									FROM TN_CTR_SLOT
//									WHERE (lane = ? AND floor > ?)
//								) b
//							)
//						)
//					)
//			AND s.lane = ?
//			`
//
//	result, err := tx.Exec(query, floor, lane, floor, floor, floor, lane, floor, lane, floor, lane)
//	//result, err := s.DB.Exec(query, floor, lane, floor, floor, floor, lane, floor, lane, floor, lane)
//	if err != nil {
//		return nil, err
//	}
//
//	affected, err := result.RowsAffected()
//
//	if err != nil {
//		return nil, err
//	}
//
//	if affected == 0 {
//		return nil, errors.New("NOT FOUND")
//	}
//
//	return result, nil
//}
//
//func UpdateOutputSlotListKeepCnt(tx *sql.Tx, itemHeight, lane, floor int) (sql.Result, error) {
//	var minStorageSlot = (floor - itemHeight + 1)
//
//	query := `
//				UPDATE TN_CTR_SLOT s
//				SET s.slot_keep_cnt = (s.slot_keep_cnt -
//											(IFNULL(
//														(
//															SELECT * FROM (
//																SELECT MAX(floor)
//																FROM TN_CTR_SLOT
//																WHERE (lane = ?) AND (FLOOR < ? AND slot_keep_cnt = 0)
//															) a
//														),
//															0
//													)
//											)
//										)
//				WHERE (s.floor >= ? AND s.floor <= ?)
//				AND s.lane = ?
//			`
//
//	result, err := tx.Exec(query, lane, floor, minStorageSlot, floor, lane)
//	//result, err := s.DB.Exec(query, lane, floor, minStorageSlot, floor, lane)
//	if err != nil {
//		return nil, err
//	}
//
//	affected, err := result.RowsAffected()
//
//	if err != nil {
//		return nil, err
//	}
//
//	if affected == 0 {
//		return nil, errors.New("NOT FOUND")
//	}
//
//	return result, nil
//}
