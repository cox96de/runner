package mock

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestNewMockDB(t *testing.T) {
	db := NewMockDB(t)
	_, err := db.CreatePipeline(context.Background())
	assert.NilError(t, err)
}
