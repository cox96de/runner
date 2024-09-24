package db

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestClient_ToMigrateSQL(t *testing.T) {
	db := NewMockDB(t)
	s, err := db.ToMigrateSQL()
	assert.NilError(t, err)
	t.Logf("%+v", s)
}
