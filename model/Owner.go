package model

import (
	log "github.com/sirupsen/logrus"
)

type OwnerInfo struct {
	Address string
}

type Owner struct {
	OwnerId  int64  `json:"owner_id"`
	PhoneNum string `json:"phone_num"`
	Address  string `json:"address"`
}

type OwnerCreateRequest struct {
	PhoneNum string `json:"phoneNum"`
	Address  string `json:"address"`
	Password string `json:"password"`
}

type OwnerUpdateRequest struct {
	PhoneNum string `json:"phoneNum"`
	Address  string `json:"address"`
}

type OwnerPwdRequest struct {
	Address  string `json:"address"`
	Password string `json:"password"`
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

func SelectPasswordByItemId(itemId int64) (string, error) {
	query := `
		SELECT password
		FROM TN_INF_OWNER o
			JOIN TN_CTR_ITEM i 
			    ON o.owner_id = i.owner_id 
		WHERE i.item_id = ?
		`

	var password string

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

func SelectOwnerList() ([]Owner, error) {
	query := `
		SELECT owner_id, phone_num, address
		FROM TN_INF_OWNER
		ORDER BY address
	`

	var ownerList []Owner

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var owner Owner
		err := rows.Scan(&owner.OwnerId, &owner.PhoneNum, &owner.Address)
		if err != nil {
			return nil, err
		}
		ownerList = append(ownerList, owner)
	}
	if err != nil {
		return nil, err
	} else {
		return ownerList, nil
	}
}

func SelectOwnerDetail(address interface{}) (Owner, error) {
	query := `
		SELECT owner_id, phone_num, address
		FROM TN_INF_OWNER
		WHERE address = ?
	`

	var owner Owner

	row := DB.QueryRow(query, address)
	err := row.Scan(&owner.OwnerId, &owner.PhoneNum, &owner.Address)

	if err != nil {
		log.Error(err)
		return owner, err
	}
	return owner, nil
}

func InsertOwner(ownerCreateRequest OwnerCreateRequest) (int64, error) {
	query := `INSERT INTO TN_INF_OWNER(
							phone_num, 
							address, 
							password)
				VALUES(?, ?, ?)
				`

	result, err := DB.Exec(query, ownerCreateRequest.PhoneNum, ownerCreateRequest.Address, ownerCreateRequest.Password)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return id, nil
}

func SelectExistOwner(address string) (int, error) {
	query := `
		select IF((SELECT owner_id FROM TN_INF_OWNER WHERE address = ?) IS NULL, 0, 1);
		`
	var exists int

	row := DB.QueryRow(query, address)
	err := row.Scan(&exists)
	if err != nil {
		log.Error(err)
		return exists, err
	}
	return exists, nil
}

func SelectOwnerAddressList() ([]Owner, error) {
	query := `
		SELECT owner_id, address
		FROM TN_INF_OWNER
	`

	var ownerList []Owner

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var owner Owner
		err := rows.Scan(&owner.OwnerId, &owner.Address)
		if err != nil {
			return nil, err
		}
		ownerList = append(ownerList, owner)
	}
	if err != nil {
		return nil, err
	} else {
		return ownerList, nil
	}
}

func UpdateOwnerInfo(ownerUpdateRequest OwnerUpdateRequest) (int64, error) {
	log.Println("UpdateOwnerInfo", ownerUpdateRequest)
	query := `UPDATE TN_INF_OWNER
				SET 
					phone_num = ?,
					address = ?
				WHERE address = ?
			`

	result, err := DB.Exec(query, ownerUpdateRequest.PhoneNum, ownerUpdateRequest.Address, ownerUpdateRequest.Address)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func ResetOwnerPassword(ownerPwdRequest OwnerPwdRequest) (int64, error) {
	query := `UPDATE TN_INF_OWNER
				SET 
					password = ?
				WHERE address = ?
			`

	result, err := DB.Exec(query, ownerPwdRequest.Password, ownerPwdRequest.Address)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
