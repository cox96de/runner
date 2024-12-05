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
}

// NewClient creates a new database client.
func NewClient(conn *gorm.DB) *Client {
	return &Client{conn: conn}
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
