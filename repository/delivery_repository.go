package repository

import (
	"APCS/data/request"
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

func (d *DeliveryRepository) SelectDeliveryByDeliveryInfo(req request.DeliveryCreateRequest) (*response.DeliveryReadResponse, error) {
	var Resp response.DeliveryReadResponse

	query := `SELECT delivery_id, delivery_name, phone_num, delivery_company
			FROM TN_INF_DELIVERY
			WHERE (delivery_name = ? and phone_num = ? and delivery_company = ?)
			`

	err := d.DB.QueryRow(query, req.DeliveryName, req.PhoneNum, req.DeliveryCompany).Scan(&Resp.DeliveryId, &Resp.DeliveryName, &Resp.PhoneNum, &Resp.DeliveryCompany)

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
