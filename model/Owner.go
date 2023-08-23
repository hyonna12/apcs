package model

import log "github.com/sirupsen/logrus"

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

	row := db.QueryRow(query, address)
	err := row.Scan(&ownerId)
	if err != nil {
		log.Error(err)
		return ownerId, err
	}
	return ownerId, nil
}
