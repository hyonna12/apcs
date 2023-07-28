package repository

import (
	"APCS/data/request"
	"APCS/data/response"
	"database/sql"
	"errors"
)

type SlotRepository struct {
	DB *sql.DB
}

func (s *SlotRepository) AssignDB(db *sql.DB) {
	s.DB = db
}

func (s *SlotRepository) SelectSlotList() (*[]response.SlotReadResponse, error) {
	var Resps []response.SlotReadResponse

	query := `
			SELECT slot_id, lane, floor, slot_keep_cnt, tray_id, item_id
			FROM TN_CTR_SLOT
			`

	rows, err := s.DB.Query(query)

	for rows.Next() {
		var Resp response.SlotReadResponse
		rows.Scan(&Resp.SlotId, &Resp.Lane, &Resp.Floor, &Resp.SlotKeepCnt, &Resp.TrayId, &Resp.ItemId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}

func (s *SlotRepository) SelectItemLocationByItemId(itemId int) (*response.SlotReadResponse, error) {
	var Resp response.SlotReadResponse
	query := `
			SELECT item_id, slot_id, lane, floor 
			FROM TN_CTR_SLOT
			WHERE item_id = ?
			`
	err := s.DB.QueryRow(query, itemId).Scan(&Resp.ItemId, &Resp.SlotId, &Resp.Lane, &Resp.Floor)

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
func (s *SlotRepository) SelectAvailableSlotList(itemHeight int) ([]response.SlotReadResponse, error) {
	slotInterval := 1
	var available = itemHeight / slotInterval
	var Resps []response.SlotReadResponse

	query := `
			SELECT slot_id, lane, floor, transport_distance, tray_id 
			FROM TN_CTR_SLOT
			WHERE (slot_enabled = 1 AND slot_keep_cnt >= ?)
			`

	rows, err := s.DB.Query(query, available)

	for rows.Next() {
		var Resp response.SlotReadResponse
		rows.Scan(&Resp.SlotId, &Resp.Lane, &Resp.Floor, &Resp.TransportDistance, &Resp.TrayId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return Resps, nil
	}
}

func (s *SlotRepository) SelectSlotListWithoutItem() (*[]response.SlotReadResponse, error) {
	var Resps []response.SlotReadResponse
	query := `
			SELECT slot_id, lane, floor, slot_keep_cnt, tray_id
			FROM TN_CTR_SLOT
			WHERE slot_enabled = 1
			`

	rows, err := s.DB.Query(query)

	for rows.Next() {
		var Resp response.SlotReadResponse
		rows.Scan(&Resp.SlotId, &Resp.Lane, &Resp.Floor, &Resp.SlotKeepCnt, &Resp.TrayId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}

func (s *SlotRepository) SelectEmptySlotList() ([]response.SlotReadResponse, error) {
	var Resps []response.SlotReadResponse
	query := `
			SELECT slot_id, lane, floor, transport_distance, slot_enabled, slot_keep_cnt, tray_id, item_id
			FROM TN_CTR_SLOT
			WHERE (slot_enabled = 1 and tray_id is null)
			`

	rows, err := s.DB.Query(query)
	for rows.Next() {
		var Resp response.SlotReadResponse
		rows.Scan(&Resp.SlotId, &Resp.Lane, &Resp.Floor, &Resp.TransportDistance, &Resp.SlotEnabled, &Resp.SlotKeepCnt, &Resp.TrayId, &Resp.ItemId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return Resps, nil
	}
}

func (s *SlotRepository) SelectSlotListForEmptyTray() ([]response.SlotReadResponse, error) {
	var Resps []response.SlotReadResponse

	query := `
			SELECT s.slot_id, s.lane, s.floor
			FROM TN_CTR_SLOT s
			JOIN TN_CTR_TRAY t
			ON s.tray_id = t.tray_id
			WHERE t.tray_occupied = 1
			`

	rows, err := s.DB.Query(query)

	for rows.Next() {
		var Resp response.SlotReadResponse
		rows.Scan(&Resp.SlotId, &Resp.Lane, &Resp.Floor, &Resp.SlotKeepCnt, &Resp.TrayId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return Resps, nil
	}
}

func (s *SlotRepository) UpdateSlot(resq request.SlotUpdateRequest) (sql.Result, error) {

	query := `
			UPDATE TN_CTR_SLOT
			SET slot_enabled = ?, slot_keep_cnt = ?, tray_id = ?, item_id = ?
			WHERE (lane = ? AND floor = ?)
			`

	result, err := s.DB.Exec(query, resq.SlotEnabled, resq.SlotKeepCnt, resq.TrayId, resq.ItemId, resq.Lane, resq.Floor)

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

func (s *SlotRepository) UpdateSlotTrayInfo(lane, floor, tray_id int) (sql.Result, error) {

	query := `
		UPDATE TN_CTR_SLOT
		SET tray_id = ?
		WHERE (lane = ? AND floor = ?)
		`

	result, err := s.DB.Exec(query, tray_id, lane, floor)

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
func (s *SlotRepository) UpdateSlotToEmptyTray(lane, floor int) (sql.Result, error) {

	query := `
		UPDATE TN_CTR_SLOT
		SET tray_id = null
		WHERE (lane = ? AND floor = ?)
		`

	result, err := s.DB.Exec(query, lane, floor)

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
func (s *SlotRepository) UpdateSlotItemInfo(lane, floor, item_id int) (sql.Result, error) {

	query := `
			UPDATE TN_CTR_SLOT
			SET item_id = ?
			WHERE (lane = ? AND floor = ?)
			`

	result, err := s.DB.Exec(query, item_id, lane, floor)

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

func (s *SlotRepository) UpdateStorageSlotList(itemHeight int, req request.SlotUpdateRequest) (sql.Result, error) {
	var minStorageSlot = (req.Floor - itemHeight + 1)
	query := `
			UPDATE TN_CTR_SLOT
			SET slot_enabled = ?, slot_keep_cnt = ?, item_id = ?
			WHERE (lane = ?) AND (floor >= ? AND floor <= ?) 
			`
	result, err := s.DB.Exec(query, req.SlotEnabled, req.SlotKeepCnt, req.ItemId, req.Lane, minStorageSlot, req.Floor)

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

func (s *SlotRepository) UpdateOutputSlotList(itemHeight int, req request.SlotUpdateRequest) (sql.Result, error) {
	var minStorageSlot = (req.Floor - itemHeight + 1)
	query := `
			UPDATE TN_CTR_SLOT
			SET slot_enabled = ?, slot_keep_cnt = floor, item_id = null
			WHERE (lane = ?) AND (floor >= ? AND floor <= ?) 
			`
	result, err := s.DB.Exec(query, req.SlotEnabled, req.Lane, minStorageSlot, req.Floor)
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

func (s *SlotRepository) SelectStorageSlotListWithTray(itemHeight, lane, floor int) ([]response.SlotReadResponse, error) {
	var minStorageSlot = (floor - itemHeight + 1)
	var Resps []response.SlotReadResponse

	query := `
			SELECT slot_id, lane, floor, tray_id
			FROM TN_CTR_SLOT 
			WHERE (lane = ?) AND (floor >= ? AND floor <= ?) AND (tray_id IS NOT NULL)
			`

	rows, err := s.DB.Query(query, lane, minStorageSlot, floor)

	for rows.Next() {
		var Resp response.SlotReadResponse
		rows.Scan(&Resp.SlotId, &Resp.Lane, &Resp.Floor, &Resp.TrayId)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return Resps, nil
	}
}
func (s *SlotRepository) SelectSlotInfoByLocation(lane, floor int) (response.SlotReadResponse, error) {

	var Resp response.SlotReadResponse

	query := `
			SELECT slot_id, lane, floor, transport_distance, slot_enabled, slot_keep_cnt, tray_id, item_id
			FROM TN_CTR_SLOT 
			WHERE (lane = ?) AND (floor  =?)
			`
	err := s.DB.QueryRow(query, lane, floor).Scan(&Resp.SlotId, &Resp.Lane, &Resp.Floor, &Resp.TransportDistance, &Resp.SlotEnabled, &Resp.SlotKeepCnt, &Resp.TrayId, &Resp.ItemId)

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

func (s *SlotRepository) UpdateStorageSlotKeepCnt(lane, floor int) (sql.Result, error) {

	query := `
			UPDATE TN_CTR_SLOT s 
			SET s.slot_keep_cnt = (s.slot_keep_cnt - ?) 
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
	result, err := s.DB.Exec(query, floor, floor, lane, floor, lane, floor, lane)
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

func (s *SlotRepository) UpdateOutputSlotKeepCnt(lane, floor int) (sql.Result, error) {

	query := `
			UPDATE TN_CTR_SLOT s 
			SET s.slot_keep_cnt = (s.slot_keep_cnt + ?) 
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
	result, err := s.DB.Exec(query, floor, floor, lane, floor, lane, floor, lane)
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

func (s *SlotRepository) UpdateOutputSlotListKeepCnt(itemHeight, lane, floor int) (sql.Result, error) {
	var minStorageSlot = (floor - itemHeight + 1)

	query := `
				UPDATE TN_CTR_SLOT s 
				SET s.slot_keep_cnt = (s.slot_keep_cnt - 
											(IFNULL( 
														( 
															SELECT * FROM ( 
																SELECT MAX(floor)
																FROM TN_CTR_SLOT 
																WHERE (lane = ?) AND (FLOOR < ? AND slot_keep_cnt = 0) 
															) a 
														), 
															0
													) 
											)
										)				
				WHERE (s.floor >= ? AND s.floor <= ?)
				AND s.lane = ?
			`

	result, err := s.DB.Exec(query, lane, floor, minStorageSlot, floor, lane)
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
