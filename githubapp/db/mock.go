package db

import (
	"fmt"
	"testing"

	"github.com/cox96de/runner/util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gotest.tools/v3/assert"
)

// NewMockDB creates a new database with tables created from given models.
// It uses sqlite in memory as the database.
func NewMockDB(t *testing.T) *Client {
	t.Helper()
	file := util.RandomID("sql-mocker")
	conn, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory", file)),
		&gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 nil, // TODO: add custom logger.
		},
	)
	assert.NilError(t, err)
	err = migrateModels(conn, &Pipeline{}, &Job{})
	assert.NilError(t, err)
	return NewClient(SQLite, conn)
}
