package repository

import "database/sql"

type OwnerRepository struct {
	DB *sql.DB
}
