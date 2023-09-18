package model

import (
	log "github.com/sirupsen/logrus"
)

type Admin struct {
	AdminId  int64
	Password string
}

func SelectAdminPassword() (int, error) {
	query := `
		SELECT password FROM admin
		`
	var password int

	row := DB.QueryRow(query)
	err := row.Scan(&password)
	if err != nil {
		log.Error(err)
		return password, err
	}
	return password, nil
}
