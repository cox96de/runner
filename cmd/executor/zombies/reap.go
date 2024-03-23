package zombies

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// doReap reaps zombie processes.
func doReap() (int, error) {
	waitStatus := unix.WaitStatus(0)
	wpid, err := unix.Wait4(-1, &waitStatus, unix.WNOHANG, nil)
	if err != nil {
		if errors.Is(err, unix.ECHILD) {
			return 0, nil
		}
		return 0, err
	}
	return wpid, nil
}

// RunReap runs reaping zombie processes.
// It reaps zombie processes every interval.
func RunReap(ctx context.Context, interval time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			for {
				pid, err := doReap()
				if err != nil {
					return err
				}
				// If pid is 0, it means no more zombies.
				// Or continue to reap.
				if pid == 0 {
					break
				}
			}
		}
	}
}
