package repository

import "database/sql"

type SlotRepository struct {
	DB *sql.DB
}
