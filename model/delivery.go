package model

import (
	"time"
)

type Delivery struct {
	DeliveryId      int64     `json:"delivery_id"`
	DeliveryName    string    `json:"delivery_name"`
	PhoneNum        string    `json:"delivery_num"`
	DeliveryCompany string    `json:"delivery_company"`
	CDatetime       time.Time `json:"c_datetime"`
	UDatetime       time.Time `json:"u_datetime"`
}

func SelectDeliveryIdByCompany(deliveryCompany string) (int64, error) {

	query := `
			SELECT delivery_id
			FROM TN_INF_DELIVERY
			WHERE delivery_name = ? 
			`

	var deliveryId int64
	row := DB.QueryRow(query, deliveryCompany)
	err := row.Scan(&deliveryId)
	if err != nil {
		return deliveryId, err
	}

	return deliveryId, nil
}

func SelectDeliveryCompanyList() ([]Delivery, error) {

	query := `
			SELECT 
				delivery_id,
				delivery_company 
			FROM TN_INF_DELIVERY
		`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	var deliverys []Delivery

	for rows.Next() {
		var delivery Delivery
		err := rows.Scan(&delivery.DeliveryId, &delivery.DeliveryCompany)
		if err != nil {
			return nil, err
		}
		deliverys = append(deliverys, delivery)
	}

	return deliverys, nil
}
