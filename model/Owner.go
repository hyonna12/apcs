package model

import (
	log "github.com/sirupsen/logrus"
)

type OwnerInfo struct {
	Address string
}

type Owner struct {
	OwnerId  int64
	PhoneNum string
	Address  string
}

func SelectOwnerIdByAddress(address string) (int64, error) {
	query := `
		SELECT owner_id
		FROM TN_INF_OWNER
		WHERE address = ?
		`

	var ownerId int64

	row := DB.QueryRow(query, address)
	err := row.Scan(&ownerId)
	if err != nil {
		log.Error(err)
		return ownerId, err
	}
	return ownerId, nil
}

func SelectPasswordByItemId(itemId int64) (int, error) {
	query := `
		SELECT password
		FROM TN_INF_OWNER o
			JOIN TN_CTR_ITEM i 
			    ON o.owner_id = i.owner_id 
		WHERE i.item_id = ?
		`

	var password int

	row := DB.QueryRow(query, itemId)
	err := row.Scan(&password)
	if err != nil {
		log.Error(err)
		return password, err
	}
	return password, nil
}

func SelectAddressByOwnerId(id interface{}) (string, error) {
	query := `
		SELECT address
		FROM TN_INF_OWNER
		WHERE owner_id = ?
		`

	var address string

	row := DB.QueryRow(query, id)
	err := row.Scan(&address)
	if err != nil {
		log.Error(err)
		return address, err
	}
	return address, nil
}
