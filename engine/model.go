package engine

import (
	"context"

	"github.com/cox96de/runner/internal/executor"
)

// RunnerSpec defines a environment to run job (compile job, ci job, etc).
type RunnerSpec struct {
	ID string
}

type Engine interface {
	// Ping checks the engine is working.
	Ping(ctx context.Context) error
	// CreateRunner creates a runner by RunnerSpec.
	CreateRunner(ctx context.Context, option *RunnerSpec) (Runner, error)
}

// Runner is an environment to run job (compile job, ci job, etc).
// In most Engine, a runner presents a clean environment, such as a container, a VM, etc.
type Runner interface {
	// Start starts the runner.
	// Before Start, the runner is not ready to run job. In most case, no resources is allocated.
	Start(ctx context.Context) error
	// GetExecutor gets an executor from the runner.
	// The Executor is a client to operate in the runner such run commands, read files.
	GetExecutor(ctx context.Context) (Executor, error)
	// Stop stops the runner. All resources should be released.
	Stop(ctx context.Context) error
}

type Executor interface {
	// Ping checks the executor is bootstrapped and ready to serve.
	Ping(ctx context.Context) error
	// StartCommand starts a command.
	// `id` is the unique id of the command, and it's unique in the executor.
	// Use that id to get logs and status.
	StartCommand(ctx context.Context, id string, opt *executor.StartCommandRequest) error
}
