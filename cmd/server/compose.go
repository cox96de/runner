package main

import (
	"github.com/cox96de/runner/db"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ComposeDB(dialect string, dsn string) (*db.Client, error) {
	var (
		conn *gorm.DB
		err  error
	)
	switch db.Dialect(dialect) {
	case db.Mysql:
		conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case db.Postgres:
		conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case db.SQLite:
		conn, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		return nil, errors.Errorf("unsupported dialect: %s", dialect)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open database connection")
	}
	return db.NewClient(db.Dialect(dialect), conn), nil
}
