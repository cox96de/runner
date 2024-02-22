package dispatch

import "github.com/cox96de/runner/api"

// CheckStatus checks if can transit from old to target status.
func CheckStatus(old api.Status, target api.Status) bool {
	if old == target {
		return true
	}
	if old.IsCompleted() {
		return false
	}
	switch old {
	case api.StatusCreated:
		return target == api.StatusQueued ||
			target == api.StatusFailed ||
			target == api.StatusSkipped
	case api.StatusQueued:
		return target == api.StatusPreparing ||
			target == api.StatusCanceling ||
			target == api.StatusFailed
	case api.StatusPreparing:
		return target.IsCompleted() ||
			target == api.StatusRunning ||
			target == api.StatusCanceling
	case api.StatusRunning:
		return target.IsCompleted() ||
			target == api.StatusCanceling
	}
	return false
}
