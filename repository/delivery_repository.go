package repository

import (
	"APCS/data/response"
	"database/sql"
)

type DeliveryRepository struct {
	DB *sql.DB
}

func (d *DeliveryRepository) AssignDB(db *sql.DB) {
	d.DB = db
}

func (d *DeliveryRepository) SelectDeliveryByDeliveryId(deliveryId int) (*[]response.DeliveryReadResponse, error) {
	var Resps []response.DeliveryReadResponse

	query := `SELECT delivery_id, delivery_name, delivery_company
			FROM TN_INF_DELIVERY
			WHERE delivery_id = ?
			`

	rows, err := d.DB.Query(query, deliveryId)

	for rows.Next() {
		var Resp response.DeliveryReadResponse
		rows.Scan(&Resp.DeliveryId, &Resp.DeliveryName, &Resp.DeliveryCompany)
		Resps = append(Resps, Resp)
	}

	if err != nil {
		return nil, err
	} else {
		return &Resps, nil
	}
}
