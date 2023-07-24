package repository

import "database/sql"

type TrayRepository struct {
	DB *sql.DB
}
