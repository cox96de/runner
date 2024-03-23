package agent

import (
	"context"
	"io"
	"time"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/log"
	"github.com/cox96de/runner/util"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

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
	_, err = e.client.UpdateStepExecution(ctx, &api.UpdateStepExecutionRequest{
		StepExecutionID: step.Executions[0].ID,
		JobExecutionID:  e.job.Executions[0].ID,
		JobID:           e.job.ID,
		Status:          lo.ToPtr(api.StatusRunning),
	})
	if err != nil {
		return errors.WithMessage(err, "failed to update step execution")
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
	logCh := make(chan struct{})
	go func() {
		defer func() {
			if err := collector.Close(); err != nil {
				logger.Errorf("failed to close log collector: %v", err)
			}
		}()
		defer close(logCh)
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
				logger.Errorf("failed to get command log: %v", err)
				_ = util.Wait(ctx, time.Second)
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
	stepStatus := api.StatusSucceeded
	if processStatus.ExitCode != 0 {
		stepStatus = api.StatusFailed
	}
	select {
	case <-logCh:
		logger.Infof("log collector is successful closed")
	case <-time.After(time.Second * 5):
		logger.Warnf("log collector is not closed in time")
	}
	if _, err = e.client.UpdateStepExecution(ctx, &api.UpdateStepExecutionRequest{
		StepExecutionID: step.Executions[0].ID,
		JobExecutionID:  e.job.Executions[0].ID,
		JobID:           e.job.ID,
		Status:          &stepStatus,
		ExitCode:        lo.ToPtr(uint32(processStatus.ExitCode)),
	}); err != nil {
		// TODO: retry if failed.
		return errors.WithMessage(err, "failed to update step execution")
	}
	logger.Infof("command is completed, exit code: %+v", processStatus.ExitCode)
	return nil
}
