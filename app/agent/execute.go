package agent

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cox96de/runner/util"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/engine"
	"github.com/pkg/errors"
)

const timeoutError = util.StringError("job timeout")

type Execution struct {
	engine           engine.Engine
	job              *api.Job
	execution        *api.JobExecution
	client           api.ServerClient
	logFlushInternal time.Duration

	runner    engine.Runner
	logWriter *logCollector

	// jobTimeoutCtx is the context for the job timeout.
	// If the job has a timeout, it will be canceled when the job is done.
	// If subprocess should be responded to the job timeout, it should use this context.
	jobTimeoutCtx context.Context
}

func NewExecution(engine engine.Engine, job *api.Job, client api.ServerClient) *Execution {
	return &Execution{
		engine:           engine,
		job:              job,
		execution:        job.Executions[len(job.Executions)-1],
		client:           client,
		logFlushInternal: time.Second,
	}
}

// Execute executes the job.
func (e *Execution) Execute(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var err error
	e.logWriter = newLogCollector(e.client, e.execution, "_", log.ExtractLogger(ctx), e.logFlushInternal)
	logger := log.ExtractLogger(ctx).WithOutput(io.MultiWriter(os.Stdout, e.logWriter))
	ctx = log.WithLogger(ctx, logger)
	if err = e.updateStatus(ctx, api.StatusPreparing); err != nil {
		return errors.WithMessage(err, "failed to update status")
	}
	e.runner, err = e.engine.CreateRunner(ctx, e, e.job)
	if err != nil {
		return err
	}
	err = e.runner.Start(ctx)
	if err != nil {
		e.stop(ctx)
		return err
	}
	defer e.stop(ctx)
	if err = e.updateStatus(ctx, api.StatusRunning); err != nil {
		return errors.WithMessage(err, "failed to update status")
	}
	if e.job.Timeout > 0 {
		var timeoutCancel context.CancelFunc
		e.jobTimeoutCtx, timeoutCancel = context.WithTimeoutCause(ctx, time.Duration(e.job.Timeout)*time.Second, timeoutError)
		defer timeoutCancel()
	} else {
		e.jobTimeoutCtx = ctx
	}
	if err = e.executeSteps(ctx); err != nil {
		if !isErrorContextCancel(err) {
			return errors.WithMessage(err, "failed to execute steps")
		}
		logger.Infof("step is aborted by timeout")
	}
	// TODO: calculate status.
	if err = e.updateStatus(ctx, api.StatusSucceeded); err != nil {
		return errors.WithMessage(err, "failed to update status")
	}
	return nil
}

func (e *Execution) updateStatus(ctx context.Context, status api.Status) error {
	execution, err := e.client.UpdateJobExecution(ctx, &api.UpdateJobExecutionRequest{
		JobID:          e.job.ID,
		JobExecutionID: e.execution.ID,
		Status:         &status,
	})
	if err != nil {
		return errors.WithMessagef(err, "failed to update job execution to %s", status)
	}
	e.execution = execution.JobExecution
	return nil
}

func (e *Execution) stop(loggerCtx context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	if err := e.runner.Stop(ctx); err != nil {
		log.ExtractLogger(loggerCtx).Errorf("failed to stop runner, the environment migh be leak: %v", err)
	}
}

func (e *Execution) executeSteps(ctx context.Context) error {
	for _, step := range e.job.Steps {
		err := e.executeStep(ctx, step)
		if err != nil {
			return errors.WithMessagef(err, "failed to execute step '%s'", step.Name)
		}
	}
	return nil
}

func (e *Execution) getExecutor(ctx context.Context, runner engine.Runner, step *api.Step) (executorpb.ExecutorClient, error) {
	if multipleContainerRunner, ok := runner.(engine.MultipleContainerRunner); ok {
		if step.Container != "" {
			return multipleContainerRunner.GetContainerExecutor(ctx, step.Container)
		}
	}
	return runner.GetExecutor(ctx)
}
