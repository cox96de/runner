package mock

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_logProvider_CreateLogWriter(t *testing.T) {
	provider := NewNopLogProvider()
	_, err := provider.CreateLogWriter(context.Background(), "test").Write([]byte("test"))
	assert.NilError(t, err)
	_, err = provider.GetDefaultLogWriter(context.Background()).Write([]byte("test"))
	assert.NilError(t, err)
}
