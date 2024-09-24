package mock

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cox96de/runner/external/redis"
	goredis "github.com/redis/go-redis/v9"
	"gotest.tools/v3/assert"
)

// NewMockRedis returns a mock redis client for testing.
func NewMockRedis(t *testing.T) *redis.Client {
	t.Helper()
	minir := miniredis.NewMiniRedis()
	err := minir.Start()
	assert.NilError(t, err)
	t.Cleanup(func() {
		minir.Close()
	})
	conn := goredis.NewClient(&goredis.Options{
		Addr: minir.Addr(),
	})
	return redis.NewClient(conn)
}
