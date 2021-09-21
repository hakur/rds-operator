package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type ClusterManager interface {
	StartCluster(ctx context.Context) (err error)
	FindMaster(ctx context.Context) (masterDSN *DSN, err error)
	HealthyMembers(ctx context.Context) (members []*DSN)
}

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
