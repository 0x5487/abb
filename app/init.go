package app

import (
	"fmt"
	"time"

	"github.com/jasonsoft/abb/config"
	"github.com/jmoiron/sqlx"
)

var (
	_db *sqlx.DB
)

func init() {
	config := config.Config()

	// set database connection
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&multiStatements=true", config.Database.Username, config.Database.Password, config.Database.Address, config.Database.DBName)
	_db = sqlx.MustConnect("mysql", connectionString)
	_db.SetMaxIdleConns(150)
	_db.SetMaxOpenConns(300)
	_db.SetConnMaxLifetime(14400 * time.Second)
}
