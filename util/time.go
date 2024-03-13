package util

import (
	"context"
	"time"
)

// Wait waits for the given duration or context to be done.
func Wait(ctx context.Context, t time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(t):
		return nil
	}
}
