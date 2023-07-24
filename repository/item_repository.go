package repository

import "database/sql"

type ItemRepository struct {
	DB *sql.DB
}
