package model

import log "github.com/sirupsen/logrus"

type Owner struct {
	Id   int    `json:"owner_id"`
	Name string `json:"owner_name"`
}

func FindAllOwners() ([]Owner, error) {
	// DB 테스트 접근
	rows, err := db.Query("SELECT owner_id, owner_name FROM TN_INF_OWNER")
	if err != nil {
		log.Error(err)
	}

	var owners []Owner

	for rows.Next() {
		var owner Owner

		if err = rows.Scan(&owner.Id, &owner.Name); err != nil {
			log.Error(err)
		}

		owners = append(owners, owner)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return owners, nil
}
