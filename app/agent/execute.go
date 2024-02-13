package agent

import (
	"context"
	"io"
	"time"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/entity"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Execution struct {
	engine engine.Engine
	job    *entity.Job

	runner engine.Runner
}

func NewExecution(engine engine.Engine, job *entity.Job) *Execution {
	return &Execution{engine: engine, job: job}
}

// Execute executes the job.
func (e *Execution) Execute(ctx context.Context) error {
	var err error
	e.runner, err = e.engine.CreateRunner(ctx, e.job)
	if err != nil {
		return err
	}
	err = e.runner.Start(ctx)
	if err != nil {
		e.stop()
		return err
	}
	defer e.stop()
	if err = e.executeSteps(ctx); err != nil {
		return errors.WithMessage(err, "failed to execute steps")
	}
	return nil
}

func (e *Execution) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	if err := e.runner.Stop(ctx); err != nil {
		log.Errorf("failed to stop runner, the environment migh be leak: %v", err)
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

func (e *Execution) executeStep(ctx context.Context, step *entity.Step) error {
	executor, err := e.runner.GetExecutor(ctx, step.Name)
	if err != nil {
		return err
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
	log.Infof("success to start command with pid: %d", startCommandResponse.Status.Pid)
	getCommandLogResp, err := executor.GetCommandLog(ctx, &executorpb.GetCommandLogRequest{
		Pid: startCommandResponse.Status.Pid,
	})
	if err != nil {
		return errors.WithMessage(err, "failed to get command log")
	}
	go func() {
		for {
			commandLog, err := getCommandLogResp.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				continue
			}
			// TODO: write log to io.Writer.
			log.Debugf("command log: %v", commandLog)
		}
	}()
	var processStatus *executorpb.ProcessStatus
	for {
		commandResponse, err := executor.WaitCommand(ctx, &executorpb.WaitCommandRequest{
			Pid:     startCommandResponse.Status.Pid,
			Timeout: int64(time.Hour), // TODO: change it, it should refer to step timeout.
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
	log.Infof("command is completed, exit code: %+v", processStatus.ExitCode)
	return nil
}
