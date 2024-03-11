package mock

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestNewMockRedis(t *testing.T) {
	mockRedis := NewMockRedis(t)
	err := mockRedis.Set(context.Background(), "key", "value", 0).Err()
	assert.NilError(t, err)
}
