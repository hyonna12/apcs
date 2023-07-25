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

func (d *DeliveryRepository) FindAll() (*[]response.DeliveryReadResponse, error) {
	var Resps []response.DeliveryReadResponse

	query := `SELECT delivery_id, delivery_name, delivery_company
				FROM TN_INF_DELIVERY`
	rows, err := d.DB.Query(query)

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

func (d *DeliveryRepository) FindById(delivertId int) (*response.DeliveryReadResponse, error) {
	var Resp response.DeliveryReadResponse

	query := `
			SELECT delivery_id, delivery_name, delivery_company
			FROM TN_INF_DELIVERY
			WHERE delivery_id = ?
			`
	err := d.DB.QueryRow(query, delivertId).Scan(&Resp.DeliveryId, &Resp.DeliveryName, &Resp.DeliveryCompany)

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

func (d *DeliveryRepository) InsertDelivery(resq request.DeliveryCreateRequest) (sql.Result, error) {

	query := `
			INSERT INTO TN_INF_DELIVERY(delivery_name, phone_num, delivery_company, c_datetime, u_datetime)
			VALUES(?, ?, ?, now(), now())
			`
	result, err := d.DB.Exec(query, resq.DeliveryName, resq.PhoneNum, resq.DeliveryCompany)

	if err != nil {
		return nil, err
	}

	return result, nil
}
