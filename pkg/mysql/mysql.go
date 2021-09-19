package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DSN struct {
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
}

func NewDBFromDSN(opts *DSN) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.DBName,
	))
	return
}
