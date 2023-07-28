package repository

import (
	"APCS/data/request"
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

func (o *OwnerRepository) SelectOwnerByOwnerInfo(req request.OwnerReadRequest) (response.OwnerReadResponse, error) {
	var Resp response.OwnerReadResponse
	query := `SELECT owner_id, owner_name, phone_num, address
			FROM TN_INF_OWNER
			WHERE owner_name = ? AND phone_num = ? AND address = ?
			`

	err := o.DB.QueryRow(query, req.OwnerName, req.PhoneNum, req.Address).Scan(&Resp.OwnerId, &Resp.OwnerName, &Resp.PhoneNum, &Resp.Address)
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
