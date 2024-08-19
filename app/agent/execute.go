package agent

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cox96de/runner/app/agent/dag"

	"github.com/cox96de/runner/lib"
	"github.com/samber/lo"

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
	jobExecution     *api.JobExecution
	stepExecutions   map[int64]*api.StepExecution
	client           api.ServerClient
	logFlushInternal time.Duration

	runner    engine.Runner
	dag       *lib.DAG[*dagNode]
	logWriter *logCollector

	// jobTimeoutCtx is the context for the job timeout.
	// If the job has a timeout, it will be canceled when the job is done.
	// If subprocess should be responded to the job timeout, it should use this context.
	jobTimeoutCtx context.Context
}

func NewExecution(engine engine.Engine, job *api.Job, client api.ServerClient) *Execution {
	e := &Execution{
		engine:           engine,
		job:              job,
		jobExecution:     job.Execution,
		stepExecutions:   map[int64]*api.StepExecution{},
		client:           client,
		logFlushInternal: time.Second,
	}
	for _, step := range e.jobExecution.Steps {
		e.stepExecutions[step.StepID] = step
	}
	return e
}

// Execute executes the job.
func (e *Execution) Execute(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var err error
	e.logWriter = newLogCollector(e.client, e.jobExecution, "_", log.ExtractLogger(ctx), e.logFlushInternal)
	logger := log.ExtractLogger(ctx).WithOutput(io.MultiWriter(os.Stdout, e.logWriter))
	ctx = log.WithLogger(ctx, logger)
	if err = e.updateJobStatus(ctx, api.StatusPreparing); err != nil {
		return errors.WithMessage(err, "failed to update status")
	}
	err = e.normalizeDAG()
	if err != nil {
		logger.Errorf("failed to normalize DAG: %v", err)
		// TODO: update job status to failed with reason.
		if err = e.updateJobStatus(ctx, api.StatusFailed); err != nil {
			return errors.WithMessage(err, "failed to update status")
		}
		return nil
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
	if err = e.updateJobStatus(ctx, api.StatusRunning); err != nil {
		return errors.WithMessage(err, "failed to update status")
	}
	e.jobTimeoutCtx = ctx
	if e.job.Timeout > 0 {
		var timeoutCancel context.CancelFunc
		e.jobTimeoutCtx, timeoutCancel = context.WithTimeoutCause(ctx, time.Duration(e.job.Timeout)*time.Second, timeoutError)
		defer timeoutCancel()
	}
	if err = e.executeSteps(ctx); err != nil {
		if !isErrorContextCancel(err) {
			return errors.WithMessage(err, "failed to execute steps")
		}
		logger.Infof("step is aborted")
	}
	jobStatus := e.calculateJobStatus()
	logger.Infof("job status is %s", jobStatus)
	if err = e.updateJobStatus(ctx, jobStatus); err != nil {
		return errors.WithMessage(err, "failed to update status")
	}
	return nil
}

func (e *Execution) updateJobStatus(ctx context.Context, status api.Status) error {
	execution, err := e.client.UpdateJobExecution(ctx, &api.UpdateJobExecutionRequest{
		JobExecutionID: e.jobExecution.ID,
		Status:         &status,
	})
	if err != nil {
		return errors.WithMessagef(err, "failed to update job jobExecution to %s", status)
	}
	e.jobExecution.Status = execution.JobExecution.Status
	return nil
}

func (e *Execution) calculateJobStatus() api.Status {
	status := api.StatusSucceeded
	for _, step := range e.stepExecutions {
		if !step.Status.IsCompleted() {
			// FIXME: It's bad status. Handle it.
			return e.jobExecution.Status
		}
		if step.Status == api.StatusFailed {
			return api.StatusFailed
		}
	}
	return status
}

func (e *Execution) stop(loggerCtx context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	if err := e.runner.Stop(ctx); err != nil {
		log.ExtractLogger(loggerCtx).Errorf("failed to stop runner, the environment migh be leak: %v", err)
	}
}

func (e *Execution) normalizeSerialSteps() {
	if e.isSerialSteps() {
		for i := 1; i < len(e.job.Steps); i++ {
			e.job.Steps[i].DependsOn = []string{e.job.Steps[i-1].Name}
		}
	}
}

func (e *Execution) isSerialSteps() bool {
	for _, step := range e.job.Steps {
		if len(step.DependsOn) > 0 {
			return false
		}
	}
	return true
}

func (e *Execution) executeSteps(ctx context.Context) error {
	e.normalizeSerialSteps()
	dagRunner := dag.NewRunner()
	for _, step := range e.job.Steps {
		err := e.executeStep(ctx, step)
		if err != nil {
			if !isErrorContextCancel(err) {
				return errors.WithMessagef(err, "failed to execute step '%s'", step.Name)
			}
			if err = e.updateStepExecution(ctx, step, lo.ToPtr(api.StatusFailed), nil); err != nil {
				return errors.WithMessage(err, "failed to update step jobExecution")
			}
		}
	}
	return dagRunner.Run()
}

func (e *Execution) getExecutor(ctx context.Context, runner engine.Runner, step *api.Step) (executorpb.ExecutorClient, error) {
	if multipleContainerRunner, ok := runner.(engine.MultipleContainerRunner); ok {
		if step.Container != "" {
			return multipleContainerRunner.GetContainerExecutor(ctx, step.Container)
		}
	}
	return runner.GetExecutor(ctx)
}
