package repository

import (
	"APCS/data/response"
	"database/sql"
	"errors"
)

type DeliveryRepository struct {
	DB *sql.DB
}

func (d *DeliveryRepository) AssignDB(db *sql.DB) {
	d.DB = db
}

func (d *DeliveryRepository) SelectDeliveryByDeliveryId(deliveryId int) (*response.DeliveryReadResponse, error) {
	var Resp response.DeliveryReadResponse

	query := `SELECT delivery_id, delivery_name, delivery_company
			FROM TN_INF_DELIVERY
			WHERE delivery_id = ?
			`

	err := d.DB.QueryRow(query, deliveryId).Scan(&Resp.DeliveryId, &Resp.DeliveryName, &Resp.DeliveryCompany)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("NOT FOUND")
		} else {
			return nil, err
		}
	} else {
		return &Resp, nil
	}
}
