package app

import (
	"fmt"
	"time"

	"github.com/jasonsoft/abb/config"
	"github.com/jmoiron/sqlx"
)

var (
	DBX     *sqlx.DB
	_config *config.Configuration
)

func init() {
	_config = config.Config()

	// set database database
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&multiStatements=true", _config.Database.Username, _config.Database.Password, _config.Database.Address, _config.Database.DBName)
	DBX = sqlx.MustConnect("mysql", connectionString)
	DBX.SetMaxIdleConns(150)
	DBX.SetMaxOpenConns(300)
	DBX.SetConnMaxLifetime(14400 * time.Second)
}
