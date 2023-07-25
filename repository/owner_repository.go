package repository

import (
	"APCS/data/response"
	"database/sql"
	"errors"
)

type OwnerRepository struct {
	DB *sql.DB
}

func (o *OwnerRepository) AssignDB(db *sql.DB) {
	o.DB = db
}

func (o *OwnerRepository) SelectOwnerByOwnerId(ownerId int) (*response.OwnerReadResponse, error) {
	var Resp response.OwnerReadResponse

	query := `SELECT owner_id, phone_num, address
			FROM TN_INF_OWNER
			WHERE owner_id = ?
			`

	err := o.DB.QueryRow(query, ownerId).Scan(&Resp.OwnerId, &Resp.PhoneNum, &Resp.Address)

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
