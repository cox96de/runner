package agent

import (
	"context"
	"io"
	"time"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/engine"
	"github.com/pkg/errors"
)

type Execution struct {
	engine    engine.Engine
	job       *api.Job
	execution *api.JobExecution
	client    api.ServerClient

	runner engine.Runner
}

func NewExecution(engine engine.Engine, job *api.Job, client api.ServerClient) *Execution {
	return &Execution{
		engine:    engine,
		job:       job,
		execution: job.Executions[len(job.Executions)-1],
		client:    client,
	}
}

// Execute executes the job.
func (e *Execution) Execute(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var err error
	if err = e.updateStatus(ctx, api.StatusPreparing); err != nil {
		return errors.WithMessage(err, "failed to update status")
	}
	e.runner, err = e.engine.CreateRunner(ctx, e.job)
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
	if err = e.executeSteps(ctx); err != nil {
		return errors.WithMessage(err, "failed to execute steps")
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
	e.execution = execution.Job
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
			return err
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

func (e *Execution) executeStep(ctx context.Context, step *api.Step) error {
	logger := log.ExtractLogger(ctx).WithField("step", step.Name)
	executor, err := e.getExecutor(ctx, e.runner, step)
	if err != nil {
		return errors.WithMessage(err, "failed to get executor")
	}
	commands := []string{"/bin/sh", "-c", "printf '%s' \"$RUNNER_SCRIPT\" | /bin/sh"}
	script := compileUnixScript(step.Commands)
	environment, err := executor.Environment(ctx, &executorpb.EnvironmentRequest{})
	if err != nil {
		return errors.WithMessage(err, "failed to get environment")
	}
	startCommandResponse, err := executor.StartCommand(ctx, &executorpb.StartCommandRequest{
		Commands: commands,
		Dir:      step.WorkingDirectory,
		Env:      append(environment.Environment, "RUNNER_SCRIPT="+script),
	})
	if err != nil {
		return errors.WithMessage(err, "failed to start command")
	}
	logger.Infof("success to start command with pid: %d", startCommandResponse.Status.Pid)
	getCommandLogResp, err := executor.GetCommandLog(ctx, &executorpb.GetCommandLogRequest{
		CommandID: startCommandResponse.CommandID,
	})
	if err != nil {
		return errors.WithMessage(err, "failed to get command log")
	}
	collector := newLogCollector(e.client, e.execution, step.Name, logger, time.Second)
	go func() {
		defer func() {
			if err := collector.Close(); err != nil {
				logger.Errorf("failed to close log collector: %v", err)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				logger.Warnf("context is done, stop getting command log")
				return
			default:
			}
			commandLog, err := getCommandLogResp.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				time.Sleep(time.Second)
				logger.Errorf("failed to get command log: %v", err)
				continue
			}
			_, err = collector.Write(commandLog.Output)
			if err != nil {
				logger.Errorf("failed to write log: %v", err)
			}
		}
	}()
	var processStatus *executorpb.ProcessStatus
	for {
		commandResponse, err := executor.WaitCommand(ctx, &executorpb.WaitCommandRequest{
			CommandID: startCommandResponse.CommandID,
			Timeout:   int64(time.Hour), // TODO: change it, it should refer to step timeout.
		})
		if err != nil {
			// TODO: auto retry.
			return errors.WithMessage(err, "failed to wait command")
		}
		processStatus = commandResponse.Status
		if commandResponse.Status.Exit {
			break
		}
	}
	logger.Infof("command is completed, exit code: %+v", processStatus.ExitCode)
	return nil
}
