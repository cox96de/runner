package util

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestWait(t *testing.T) {
	t.Run("sleep", func(t *testing.T) {
		start := time.Now()
		err := Wait(context.Background(), time.Millisecond*10)
		assert.NilError(t, err)
		assert.Assert(t, time.Since(start) > time.Millisecond*10)
	})
	t.Run("context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
		defer cancel()
		start := time.Now()
		err := Wait(ctx, time.Second)
		assert.Error(t, err, "context deadline exceeded")
		assert.Assert(t, time.Since(start) < time.Millisecond*100)
	})
}
