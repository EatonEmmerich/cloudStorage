package db

import (
	"database/sql"
	_ "embed"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var database = flag.String("database", "cloudStorage", "The name of the db to populate the data in")
var user = flag.String("user", "root", "The name of the db user for the app")
var password = flag.String("password", "cloudStorage", "The password of the db user specified in 'user' parameter")

func New() (*sql.DB,error) {
	dbc,err := sql.Open("mysql", (*user) + ":" + (*password) + "@/" + (*database))
	if err != nil {
		return nil, err
	}

	dbc.SetConnMaxLifetime(time.Minute * 3)
	dbc.SetMaxOpenConns(10)
	dbc.SetMaxIdleConns(10)

	return dbc, err
}

