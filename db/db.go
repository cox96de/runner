package db

import (
	"github.com/pkg/errors"
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

// Transaction executes a block of code inside a database transaction.
func (c *Client) Transaction(fc func(client *Client) error) error {
	return c.conn.Transaction(func(tx *gorm.DB) error {
		return fc(&Client{conn: tx})
	})
}

// IsRecordNotFoundError returns true if the error is a record not found error.
func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
