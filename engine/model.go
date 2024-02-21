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
	// The name typically is the name of the step.
	GetExecutor(ctx context.Context, stepName string) (executorpb.ExecutorClient, error)
	// Stop stops the runner. All resources should be released.
	Stop(ctx context.Context) error
}
