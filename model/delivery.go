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

type DeliveryReadRequest struct {
	DeliveryName    string `json:"delivery_name"`
	PhoneNum        string `json:"phone_num"`
	DeliveryCompany string `json:"company"`
}

func SelectDeliveryByDeliveryInfo(deliveryReadRequest DeliveryReadRequest) (Delivery, error) {

	query := `
			SELECT 
				delivery_id, 
				delivery_name, 
				phone_num, 
				delivery_company
			FROM TN_INF_DELIVERY
			WHERE 
			    delivery_name = ? 
			  	AND phone_num = ?
			  	AND delivery_company = ?
			`

	var delivery Delivery
	row := db.QueryRow(query, deliveryReadRequest.DeliveryName, deliveryReadRequest.PhoneNum, deliveryReadRequest.DeliveryCompany)
	err := row.Scan(&delivery.DeliveryId, &delivery.DeliveryName, &delivery.PhoneNum, &delivery.DeliveryCompany)
	if err != nil {
		return Delivery{}, err
	}

	return delivery, nil
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
