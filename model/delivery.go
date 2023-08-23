package model

import (
	"time"
)

type Delivery struct {
	DeliveryId      int64
	DeliveryName    string
	PhoneNum        string
	DeliveryCompany string
	CDatetime       time.Time
	UDatetime       time.Time
}

func SelectDeliveryIdByCompany(deliveryCompany string) (int64, error) {

	query := `
			SELECT delivery_id
			FROM TN_INF_DELIVERY
			WHERE delivery_name = ? 
			`

	var deliveryId int64
	row := db.QueryRow(query, deliveryCompany)
	err := row.Scan(&deliveryId)
	if err != nil {
		return deliveryId, err
	}

	return deliveryId, nil
}

func SelectDeliveryCompanyList() ([]string, error) {

	query := `
			SELECT delivery_company 
			FROM TN_INF_DELIVERY
		`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var deliveryCompanys []string

	for rows.Next() {
		var deliveryCompany string
		err := rows.Scan(&deliveryCompany)
		if err != nil {
			return nil, err
		}
		deliveryCompanys = append(deliveryCompanys, deliveryCompany)
	}

	return deliveryCompanys, nil
}
