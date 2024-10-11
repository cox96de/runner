package db

import (
	"github.com/cockroachdb/errors"
	"gorm.io/gorm"
)

type Dialect string

const (
	Postgres Dialect = "postgres"
	SQLite   Dialect = "sqlite"
	Mysql    Dialect = "mysql"
)

type Client struct {
	conn *gorm.DB
	// dialect is the database dialect used by the client.
	// Client uses dialect to generate specific SQL for better performance.
	dialect Dialect
}

// NewClient creates a new database client.
func NewClient(dialect Dialect, conn *gorm.DB) *Client {
	return &Client{conn: conn, dialect: dialect}
}

// IsRecordNotFoundError returns true if the error is a record not found error.
func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func migrateModels(conn *gorm.DB, models ...interface{}) error {
	err := conn.AutoMigrate(models...)
	if err != nil {
		if sqlDB, err := conn.DB(); err == nil {
			_ = sqlDB.Close()
		}
		return errors.WithMessage(err, "failed to migrate models")
	}
	return nil
}
