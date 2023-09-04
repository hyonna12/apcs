package model

import (
	"apcs_refactored/customerror"
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type SlotUpdateRequest struct {
	SlotEnabled bool
	SlotKeepCnt int
	TrayId      sql.NullInt64
	ItemId      sql.NullInt64
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
	TrayId            sql.NullInt64
	ItemId            sql.NullInt64
	CheckDatetime     time.Time
	CDatetime         time.Time
	UDatetime         time.Time
}

func SelectSlotList() ([]Slot, error) {
	query := `
		SELECT * FROM TN_CTR_SLOT
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		// &slot.SlotEnabled를 직접 스캔하면
		// couldn't convert "\x01" into type bool, wantErr false 에러 발생함.
		// 이를 우회하기 위해 임시 변수 사용
		var slotEnabled []uint8
		err := rows.Scan(
			&slot.SlotId,
			&slot.Lane,
			&slot.Floor,
			&slot.TransportDistance,
			&slotEnabled,
			&slot.SlotKeepCnt,
			&slot.TrayId,
			&slot.ItemId,
			&slot.CheckDatetime,
			&slot.CDatetime,
			&slot.UDatetime,
		)
		if err != nil {
			return nil, err
		}
		slot.SlotEnabled = slotEnabled[0] == 1
		slots = append(slots, slot)
	}

	return slots, nil
}

func SelectSlotsByItemIds(itemIds []int64) ([]Slot, error) {
	// WHERE IN 절에 들어갈 파라미터 제작
	var itemIdsStr []string
	for _, id := range itemIds {
		itemIdsStr = append(itemIdsStr, strconv.FormatInt(id, 10))
	}
	params := strings.Join(itemIdsStr, ",")

	query := `
		SELECT
			*
		FROM TN_CTR_SLOT s
		WHERE 
		    s.item_id IN ( ` + params + `)
			AND s.tray_id IS NOT NULL
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var slots []Slot

	for rows.Next() {
		var slot Slot
		var slotEnabled []uint8
		err := rows.Scan(
			&slot.SlotId,
			&slot.Lane,
			&slot.Floor,
			&slot.TransportDistance,
			&slotEnabled,
			&slot.SlotKeepCnt,
			&slot.TrayId,
			&slot.ItemId,
			&slot.CheckDatetime,
			&slot.CDatetime,
			&slot.UDatetime,
		)
		if err != nil {
			return nil, err
		}
		slot.SlotEnabled = slotEnabled[0] == 1
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
				AND lane != 1 
				AND lane != 2
			`

	rows, err := db.Query(query, itemHeight)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		err := rows.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.TransportDistance, &slot.TrayId)
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

// SelectSlotListForEmptyTray
//
// 빈 트레이를 수납할 슬롯 선정
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
			WHERE t.tray_occupied = 0
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

func SelectSlotsInLaneByItemId(itemId int64) ([]Slot, error) {
	query :=
		`
			SELECT
				*
			FROM TN_CTR_SLOT
			WHERE lane = (
				SELECT
					lane
				FROM TN_CTR_SLOT
				WHERE item_id = ?
				LIMIT 1
			)
			ORDER BY FLOOR

		`

	rows, err := db.Query(query, itemId)
	if err != nil {
		return nil, err
	}

	var slots []Slot

	for rows.Next() {
		var slot Slot
		var slotEnabled []uint8
		err := rows.Scan(
			&slot.SlotId,
			&slot.Lane,
			&slot.Floor,
			&slot.TransportDistance,
			&slotEnabled,
			&slot.SlotKeepCnt,
			&slot.TrayId,
			&slot.ItemId,
			&slot.CheckDatetime,
			&slot.CDatetime,
			&slot.UDatetime,
		)
		if err != nil {
			return nil, err
		}
		slot.SlotEnabled = slotEnabled[0] == 1
		slots = append(slots, slot)
	}

	return slots, nil
}

func UpdateSlots(slots []Slot) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	var totalAffected int64

	for _, slot := range slots {
		query := `
			UPDATE TN_CTR_SLOT
			SET
				lane = ?,
				floor = ?,
				transport_distance = ?,
				slot_enabled = ?,
				slot_keep_cnt = ?,
				tray_id = ?,
				item_id = ?,
				check_datetime = ?, 
				u_datetime = NOW()
			WHERE slot_id = ?
		`

		result, err := tx.Exec(query,
			slot.Lane,
			slot.Floor,
			slot.TransportDistance,
			slot.SlotEnabled,
			slot.SlotKeepCnt,
			slot.TrayId,
			slot.ItemId,
			slot.CheckDatetime,
			slot.SlotId,
		)
		if err != nil {
			return 0, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}

		totalAffected += affected
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return totalAffected, nil
}

func UpdateSlot(request SlotUpdateRequest) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	query := `
			UPDATE TN_CTR_SLOT
			SET 
			    slot_enabled = ?, 
			    slot_keep_cnt = ?, 
			    tray_id = ?, 
			    item_id = ?,
			    u_datetime = NOW()
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

func UpdateStorageSlotList(itemHeight int, req SlotUpdateRequest) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	var minStorageSlot = req.Floor - itemHeight + 1
	query := `
			UPDATE TN_CTR_SLOT
			SET 
			    slot_enabled = ?, 
			    slot_keep_cnt = ?, 
			    item_id = ?
			WHERE lane = ? 
				AND floor BETWEEN ? AND ?
			`

	result, err := tx.Exec(query, req.SlotEnabled, req.SlotKeepCnt, req.ItemId, req.Lane, minStorageSlot, req.Floor)
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

func UpdateOutputSlotList(itemHeight int, req SlotUpdateRequest) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	var minStorageSlot = req.Floor - itemHeight + 1
	query := `
			UPDATE TN_CTR_SLOT s
			SET 
			    s.slot_enabled = ?, 
			    s.slot_keep_cnt = (s.floor -
					(IFNULL(
								(
									SELECT * FROM (
										SELECT MAX(floor)
										FROM TN_CTR_SLOT
										WHERE (lane = ?) AND (floor < ? AND slot_keep_cnt = 0)
									) a
								), 0
							)
					)
				), 
			    s.item_id = null
			WHERE 
			    s.lane = ?
			  	AND s.floor BETWEEN ? AND ?
			`

	result, err := tx.Exec(query, req.SlotEnabled, req.Lane, minStorageSlot, req.Lane, minStorageSlot, req.Floor)
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

func SelectStorageSlotListWithTray(itemHeight, lane, floor int) ([]Slot, error) {
	var minStorageSlot = floor - itemHeight + 1
	query := `
			SELECT 
			    slot_id, 
			    lane, 
			    floor, 
			    tray_id
			FROM TN_CTR_SLOT
			WHERE 
			    lane = ? 
			  	AND floor BETWEEN ? AND ?
			  	AND tray_id IS NOT NULL
			`
	rows, err := db.Query(query, lane, minStorageSlot, floor)
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

func SelectSlotInfoByLocation(lane, floor int) (Slot, error) {

	query := `
			SELECT 
			    slot_id,
			    lane,
			    floor,
			    transport_distance,
			    slot_enabled, 
			    slot_keep_cnt,
			    tray_id,
			    item_id
			FROM TN_CTR_SLOT
			WHERE 
			    lane = ?
			  	AND floor = ?
			`

	var slot Slot

	row := db.QueryRow(query, lane, floor)
	err := row.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.TransportDistance, &slot.SlotEnabled, &slot.SlotKeepCnt, &slot.TrayId, &slot.ItemId)
	if err != nil {
		return Slot{}, err
	}

	return slot, nil
}

func UpdateStorageSlotKeepCnt(lane, floor, itemHeight int) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	query := `
			UPDATE TN_CTR_SLOT s
			SET s.slot_keep_cnt = (s.slot_keep_cnt - 
				IF((s.floor=s.slot_keep_cnt), 
					?, 
					(
						SELECT * FROM 
						(
							SELECT slot_keep_cnt 
							FROM TN_CTR_SLOT 
							WHERE (lane = ? AND FLOOR = ?)
						) c 
					)
				) 
			)
			WHERE (s.floor > ? AND s.floor <=
				IFNULL(
						(
							SELECT * FROM (
								SELECT MIN(floor) - 1
								FROM TN_CTR_SLOT
								WHERE (lane = ?) AND (floor > ? AND slot_keep_cnt = 0)
							) a
						),
						(
							SELECT * FROM (
								SELECT MAX(floor)
								FROM TN_CTR_SLOT
								WHERE (lane = ? AND floor > ?)
							) b
						)
					)
				)
			AND s.lane = ?
			`

	result, err := tx.Exec(query, floor, lane, floor, floor, lane, floor, lane, floor, lane)
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

func UpdateOutputSlotKeepCnt(lane, floor int) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// TODO - 쿼리 분리
	query := `
			UPDATE TN_CTR_SLOT s
			SET s.slot_keep_cnt = (s.slot_keep_cnt +
				IFNULL(
						(
							SELECT * FROM (
								SELECT ? - MAX(FLOOR)
								FROM TN_CTR_SLOT
								WHERE (lane = ?) AND (FLOOR < ? AND slot_keep_cnt = 0)
							) i
						),
							?
						)
					)
			WHERE (s.floor > ? AND s.floor <=
				IFNULL(
						(
							SELECT * FROM (
								SELECT MIN(floor) - 1
								FROM TN_CTR_SLOT
								WHERE (lane = ?) AND (floor > ? AND slot_keep_cnt = 0)
								) a
							),
							(
								SELECT * FROM (
									SELECT MAX(floor)
									FROM TN_CTR_SLOT
									WHERE (lane = ? AND floor > ?)
								) b
							)
						)
					)
			AND s.lane = ?
			`

	result, err := tx.Exec(query, floor, lane, floor, floor, floor, lane, floor, lane, floor, lane)
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

/* func UpdateOutputSlotListKeepCnt(itemHeight, lane, floor int) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	var minStorageSlot = floor - itemHeight + 1

	query := `
				UPDATE TN_CTR_SLOT s
				SET s.slot_keep_cnt = (s.floor -
											(IFNULL(
														(
															SELECT * FROM (
																SELECT MAX(floor)
																FROM TN_CTR_SLOT
																WHERE (lane = ?) AND (floor < ? AND slot_keep_cnt = 0)
															) a
														), 0
													)
											)
										)
				WHERE (s.floor >= ? AND s.floor <= ?)
				AND s.lane = ?
			`

	result, err := tx.Exec(query, lane, floor, minStorageSlot, floor, lane)
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
} */

func SelectSlotByItemId(item_id int64) (Slot, error) {

	query := `
			SELECT 
				slot_id, 
				lane, 
				floor, 
				tray_id, 
				item_id 
			FROM TN_CTR_SLOT
			WHERE 
				item_id = ? 
				AND tray_id is not null
			`

	var slot Slot

	row := db.QueryRow(query, item_id)
	err := row.Scan(&slot.SlotId, &slot.Lane, &slot.Floor, &slot.TrayId, &slot.ItemId)
	if err != nil {
		return Slot{}, err
	}

	return slot, nil
}

func UpdateSlotToEmptyTray(request SlotUpdateRequest) (int64, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	query := `
			UPDATE TN_CTR_SLOT
			SET tray_id = null
			WHERE 
				lane = ? 
				AND floor = ?
			`

	result, err := tx.Exec(query, request.Lane, request.Floor)
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

// SelectEmptyTray
//
// 빈 트레이 선정
// TODO - 최적화 필요
func SelectEmptyTray() (Slot, error) {
	query := `
			SELECT
			    slot_id,
			    slot_keep_cnt,
			    lane, 
			    floor, 
			    min(tray_id)
			FROM TN_CTR_SLOT
			WHERE 
			    slot_enabled = 1
				AND tray_id IS NOT null
			`

	var slot Slot

	row := db.QueryRow(query)

	err := row.Scan(
		&slot.SlotId,
		&slot.SlotKeepCnt,
		&slot.Lane,
		&slot.Floor,
		&slot.TrayId)
	if err != nil {
		log.Error(err)
		return slot, err
	}

	return slot, nil
}
