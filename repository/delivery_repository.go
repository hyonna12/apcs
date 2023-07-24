package repository

import "database/sql"

type DeliveryRepository struct {
	DB *sql.DB
}
