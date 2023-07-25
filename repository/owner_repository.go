package repository

import (
	"APCS/data/response"
	"database/sql"
)

type OwnerRepository struct {
	DB *sql.DB
}

func (o *OwnerRepository) AssignDB(db *sql.DB) {
	o.DB = db
}

func (o *OwnerRepository) SelectOwnerByOwnerId(ownerId int) (*[]response.OwnerReadResponse, error) {
	var Resps []response.OwnerReadResponse
	query := `SELECT owner_id, phone_num, address
			FROM TN_INF_OWNER
			WHERE owner_id = ?
			`

	rows, err := o.DB.Query(query, ownerId)

	for rows.Next() {
		var Resp response.OwnerReadResponse
		rows.Scan(&Resp.OwnerId, &Resp.PhoneNum, &Resp.Address)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}
