package lib

import (
	"context"
	"fmt"
	"time"
)

type Locker interface {
	// Lock tries to lock a key with the given value for a period.
	// It returns whether locking was successful and whether an error occurred.
	Lock(ctx context.Context, key, value string, expiresIn time.Duration) (bool, error)
	// Unlock tries to unlock a key. It only succeeds if the value matches.
	Unlock(ctx context.Context, key string) (bool, error)
}

// BuildJobRequestLockKey builds a lock key for a job request.
// The key is used to lock a job request to prevent multiple requests for the same job execution.
// Server keeps the lock for a period, and waits the agent to pick up the job execution.
// The job request process is a 2 phase commit process. Phase 1 is to lock the job request,
// and phase 2 is to really get the job execution.
func BuildJobRequestLockKey(jobExecutionID int64) string {
	return fmt.Sprintf("job_request:%d:lock", jobExecutionID)
}

// BuildJobExecutionLockKey builds a lock key for a job execution.
// The key is used to lock a job execution to prevent concurrent update job execution.
func BuildJobExecutionLockKey(jobExecutionID int64) string {
	return fmt.Sprintf("job_execution:%d:lock", jobExecutionID)
}
