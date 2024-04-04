package model

import (
	log "github.com/sirupsen/logrus"
)

type Admin struct {
	Password string
	IbId     int
}

func SelectAdminPassword() (string, error) {
	query := `
		SELECT password FROM admin
		`
	var password string

	row := DB.QueryRow(query)
	err := row.Scan(&password)
	if err != nil {
		log.Error(err)
		return password, err
	}
	return password, nil
}

func SelectExistPassword() (int, error) {
	query := `
		select IF((SELECT password FROM admin) IS NULL, 0, 1);
		`
	var exists int

	row := DB.QueryRow(query)
	err := row.Scan(&exists)
	if err != nil {
		log.Error(err)
		return exists, err
	}
	return exists, nil
}

func SelectIbId() (int, error) {
	query := `SELECT ib_id From admin`

	var ib_id int
	row := DB.QueryRow(query)
	err := row.Scan(&ib_id)
	if err != nil {
		log.Error(err)
		return ib_id, err
	}
	return ib_id, nil
}

func InsertAdminPwd(password interface{}) (int64, error) {
	query := `UPDATE admin
				SET password = ?
			`

	result, err := DB.Exec(query, password)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
