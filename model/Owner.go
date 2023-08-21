package model

import log "github.com/sirupsen/logrus"

type OwnerInfo struct {
	OwnerName string
	PhoneNum  string
	Address   string
}

type Owner struct {
	OwnerId   int64
	OwnerName string
	PhoneNum  string
	Address   string
}

func SelectOwnerByOwnerInfo(info OwnerInfo) (Owner, error) {
	query :=
		`SELECT 
    			owner_id, 
    			owner_name, 
    			phone_num, 
    			address
		FROM TN_INF_OWNER
		WHERE owner_name = ? 
			AND phone_num = ? 
		  	AND address = ?
		`

	var owner Owner

	row := db.QueryRow(query, info.OwnerName, info.PhoneNum, info.Address)
	err := row.Scan(&owner.OwnerId, &owner.OwnerName, &owner.PhoneNum, &owner.Address)
	if err != nil {
		log.Error(err)
		return Owner{}, err
	}

	return owner, nil
}
