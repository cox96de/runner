package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func TestClient_CreatePipeline(t *testing.T) {
	db := NewMockDB(t, &Pipeline{})
	createdPipeline, err := db.CreatePipeline(context.Background())
	require.NoError(t, err)
	assert.Assert(t, createdPipeline.ID > 0)
}
