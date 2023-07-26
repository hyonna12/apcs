package config

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

const (
	host     = "lineworldap.iptime.org"
	port     = 33306
	user     = "apcs_dev"
	password = "apcs@123"
	dbName   = "apcs_dev"
)

func DBConnection() *sql.DB {
	sqlInfo := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbName)

	db, err := sql.Open("mysql", sqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	log.Info().Msg("Connected to database!!")

	return db
}
