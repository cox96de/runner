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

func (e *Execution) updateStepExecution(ctx context.Context, step *api.Step, status *api.Status, exitCode *uint32) error {
	stepExecution, ok := e.stepExecutions[step.ID]
	if !ok {
		return errors.Errorf("step execution not found: %d", step.ID)
	}
	_, err := e.client.UpdateStepExecution(ctx, &api.UpdateStepExecutionRequest{
		StepExecutionID: stepExecution.ID,
		JobExecutionID:  e.jobExecution.ID,
		JobID:           e.job.ID,
		Status:          status,
		ExitCode:        exitCode,
	})
	if err != nil {
		return err
	}
	if status != nil {
		stepExecution.Status = *status
	}
	return nil
}

func (e *Execution) executeStep(ctx context.Context, step *api.Step) error {
	logger := log.ExtractLogger(ctx).WithField("step", step.Name)
	executor, err := e.getExecutor(ctx, e.runner, step)
	if err != nil {
		return errors.WithMessage(err, "failed to get executor")
	}
	getRuntimeInfoResp, err := executor.GetRuntimeInfo(ctx, &executorpb.GetRuntimeInfoRequest{})
	if err != nil {
		return errors.WithMessage(err, "failed to get runtime info")
	}
	var (
		commands []string
		script   string
	)
	switch getRuntimeInfoResp.OS {
	case "windows":
		commands = getWindowsCommands()
		script = compileWindowsScript(step.Commands)
	case "linux", "darwin":
		commands = getUnixCommands()
		script = compileUnixScript(step.Commands)
	default:
		return errors.Errorf("unsupported os: '%s'", getRuntimeInfoResp.OS)
	}

	environment, err := executor.Environment(ctx, &executorpb.EnvironmentRequest{})
	if err != nil {
		return errors.WithMessage(err, "failed to get environment")
	}
	err = e.updateStepExecution(ctx, step, lo.ToPtr(api.StatusRunning), nil)
	if err != nil {
		return errors.WithMessage(err, "failed to update step jobExecution")
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
	collector := newLogCollector(e.client, e.jobExecution, step.Name, logger, e.logFlushInternal)
	logCh := make(chan struct{})
	go func() {
		defer func() {
			if err := collector.Close(); err != nil {
				logger.Errorf("failed to close log collector: %v", err)
			}
			close(logCh)
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
		commandResponse, err := executor.WaitCommand(e.jobTimeoutCtx, &executorpb.WaitCommandRequest{
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

	if err = e.updateStepExecution(ctx, step, &stepStatus, lo.ToPtr(uint32(processStatus.ExitCode))); err != nil {
		// TODO: retry if failed.
		return errors.WithMessage(err, "failed to update step jobExecution")
	}
	logger.Infof("command is completed, exit code: %+v", processStatus.ExitCode)
	return nil
}
