package config

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
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

	log.Info("Connected to database!!")

	return db
}
