package model

import "database/sql"

type TrayBuffer struct {
	Count      int          `json:"count"`
	Updatetime sql.NullTime `json:"updatetime"`
}

func SelectTrayBufferState() (TrayBuffer, error) {

	query := `SELECT 
				tray_count, 
				updatetime
			FROM tray_buffer
			ORDER BY updatetime DESC LIMIT 1
			`

	var trayBuffer TrayBuffer
	row := DB.QueryRow(query)
	err := row.Scan(
		&trayBuffer.Count,
		&trayBuffer.Updatetime,
	)
	if err != nil {
		return TrayBuffer{}, err
	}

	return trayBuffer, nil
}

func InsertBufferState(count int) error {
	query := `INSERT INTO tray_buffer(tray_count, updatetime)
			VALUES(?, NOW())
			`

	_, err := DB.Exec(query, count)
	if err != nil {
		return err
	}

	return nil
}
