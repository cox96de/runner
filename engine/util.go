package engine

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// WaitExecutorReady waits for the executor to be ready.
// It will block until the executor is ready or the context is done.
func WaitExecutorReady(ctx context.Context, c Executor, interval time.Duration, timeout time.Duration) error {
	if interval == 0 {
		return errors.Errorf("interval must be greater than 0")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := c.Ping(ctx)
	if err == nil {
		return nil
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := c.Ping(ctx)
			if err == nil {
				return nil
			}
		}
	}
}
