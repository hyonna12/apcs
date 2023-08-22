package model

import (
	"apcs_refactored/config"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

var (
	db *sql.DB
)

func InitDB() {
	log.Info("Started initializing database connection")
	dbConfig := config.Config.Database
	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DbName)

	connection, err := sql.Open(dbConfig.DriverName, dbUrl)
	if err != nil {
		log.Panic("Failed to connect to database")
	}

	if err := connection.Ping(); err != nil {
		log.Panic("Failed to connect to database")
	}

	log.Info("Successfully initialized database connection")

	db = connection
}

func CloseDB() {
	err := db.Close()
	if err != nil {
		log.Panicf("Failed to close database")
	}

	log.Info("Successfully closed database connection")
}
