package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDB new *gorm.DB from mysql DSN
func NewDB(dsn string) (db *gorm.DB, err error) {
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	return
}

// MysqlUserTable mysql user table data fields
type MysqlUserTable struct {
	User     string `gorm:"user"`
	Password string `gorm:"password"`
	Host     string `gorm:"host"`
}
