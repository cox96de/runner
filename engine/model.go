package engine

import (
	"context"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/app/executor/executorpb"
)

//go:generate mockgen -destination mockgen_test.go -source model.go -package engine . Engine,Runner,Executor

type Engine interface {
	// Ping checks the engine is working.
	Ping(ctx context.Context) error
	// CreateRunner creates a runner by RunnerSpec.
	CreateRunner(ctx context.Context, option *api.Job) (Runner, error)
}

// Runner is an environment to run job (compile job, ci job, etc).
// In most Engine, a runner presents a clean environment, such as a container, a VM, etc.
type Runner interface {
	// Start starts the runner.
	// Before Start, the runner is not ready to run job. In most case, no resources is allocated.
	Start(ctx context.Context) error
	// GetExecutor gets an executor from the runner.
	// The Executor is a client to operate in the runner such run commands, read files.
	GetExecutor(ctx context.Context) (executorpb.ExecutorClient, error)
	// Stop stops the runner. All resources should be released.
	Stop(ctx context.Context) error
}

// MultipleContainerRunner is a runner that has multiple containers.
type MultipleContainerRunner interface {
	Runner
	// GetContainerExecutor gets an executor from the runner.
	// The Executor is a client to operate in the runner such run commands, read files.
	// The containerName is the name of the container to run the executor.
	// For non-container runner, the containerName can be ignored.
	GetContainerExecutor(ctx context.Context, containerName string) (executorpb.ExecutorClient, error)
}
