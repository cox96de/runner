package agent

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/log"
	"github.com/cox96de/runner/util"
	"github.com/samber/lo"
)

func (e *Execution) updateStepExecution(ctx context.Context, step *api.Step, status *api.Status, exitCode *uint32) error {
	stepExecution, ok := e.stepExecutions[step.ID]
	if !ok {
		return errors.Errorf("step execution not found: %d", step.ID)
	}
	_, err := e.client.UpdateStepExecution(ctx, &api.UpdateStepExecutionRequest{
		StepExecutionID: stepExecution.ID,
		Status:          status,
		ExitCode:        exitCode,
	})
	if err != nil {
		return err
	}
	if status != nil {
		stepExecution.Status = *status
	}
	if exitCode != nil {
		stepExecution.ExitCode = *exitCode
	}
	return nil
}

func (e *Execution) executeStep(ctx context.Context, step *api.Step) (err error) {
	logger := log.ExtractLogger(ctx).WithField("step", step.Name)
	collector := newLogCollector(e.client, e.jobExecution, step.Name, logger, e.logFlushInternal)
	defer func() {
		if err == nil {
			return
		}
		if _, writeErr := collector.Write([]byte("$$ Internal Error: " + err.Error())); writeErr != nil {
			logger.Errorf("failed to write error log: %v", writeErr)
		}
		if closeErr := collector.Close(); closeErr != nil {
			logger.Errorf("failed to close log collector: %v", closeErr)
		}
		// TODO: update to failed.
	}()
	continueExecute, err := e.preStep(step)
	if err != nil {
		// TODO: update to failed.
		return errors.WithMessage(err, "failed to pre step")
	}
	if !continueExecute {
		logger.Infof("skip step")
		if err := e.updateStepExecution(ctx, step, lo.ToPtr(api.StatusSkipped), nil); err != nil {
			return errors.WithMessage(err, "failed to update step jobExecution")
		}
		return nil
	}
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
	var (
		stepStatus = api.StatusSucceeded
		exitCode   int32
	)
	for {
		commandResponse, err := executor.WaitCommand(e.jobCtx, &executorpb.WaitCommandRequest{
			CommandID: startCommandResponse.CommandID,
			Timeout:   int64(time.Hour), // TODO: change it, it should refer to step timeout.
		})
		if err != nil {
			statusError, ok := status.FromError(err)
			if !ok {
				// TODO: auto retry.
				return errors.WithMessage(err, "failed to wait command")
			}
			if statusError.Code() == codes.Canceled {
				// Aborted by context.
				logger.Infof("context is canceled, stop waiting command")
				stepStatus = api.StatusFailed
				break
			}
			return errors.WithMessage(err, "failed to wait command")
		}
		processStatus := commandResponse.Status
		if commandResponse.Status.Exit {
			if processStatus.ExitCode != 0 {
				stepStatus = api.StatusFailed
				exitCode = processStatus.ExitCode
			}
			break
		}
	}

	select {
	case <-logCh:
		logger.Infof("log collector is successful closed")
	case <-time.After(time.Second * 5):
		logger.Warnf("log collector is not closed in time")
	}
	if err = e.updateStepExecution(ctx, step, &stepStatus, lo.ToPtr(uint32(exitCode))); err != nil {
		// TODO: retry if failed.
		return errors.WithMessage(err, "failed to update step jobExecution")
	}
	logger.Infof("command is completed, exit code: %+v", exitCode)
	return nil
}

func (e *Execution) continueWhenNoPreFailed(step *api.Step) (bool, error) {
	deepPres, err := e.dag.DeepPre(step.Name)
	if err != nil {
		return false, errors.WithMessage(err, "failed to get deep previous steps")
	}
	for _, pre := range deepPres {
		if e.stepExecutions[pre.Step.ID].Status == api.StatusFailed {
			return false, nil
		}
	}
	return true, nil
}

func (e *Execution) preStep(step *api.Step) (bool, error) {
	if abortedReason(e.abortedReason.Load()) != None {
		return false, nil
	}
	return e.continueWhenNoPreFailed(step)
}
