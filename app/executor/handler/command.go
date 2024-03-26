package handler

import (
	"context"
	"io"
	"os/exec"
	"time"

	"github.com/cox96de/runner/util"

	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/lib"
	"github.com/pkg/errors"
)

const (
	defaultRingBufferSize = 1 << 20
	flushLogInterval      = time.Millisecond * 100
)

var randomStringFunc = util.RandomString

func (h *Handler) GetCommandLog(request *executorpb.GetCommandLogRequest, server executorpb.Executor_GetCommandLogServer) error {
	h.commandLock.RLock()
	c, ok := h.commands[request.CommandID]
	h.commandLock.RUnlock()
	if !ok {
		return errors.Errorf("command with pid %s not found", request.GetCommandID())
	}
	bufSize := 1024
	logBuf := make([]byte, bufSize)
	ctx := server.Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:

		}
		n, readErr := c.logWriter.Read(logBuf)
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return errors.WithMessage(readErr, "failed to read from log buffer")
		}
		sendErr := server.Send(&executorpb.Log{
			Output: logBuf[:n],
		})
		if sendErr != nil {
			return errors.WithMessage(sendErr, "failed to send log")
		}
		if n < bufSize {
			// Cool down a bit.
			_ = util.Wait(ctx, flushLogInterval)
		}
	}
}

func (h *Handler) setCommand(c *command) (string, error) {
	h.commandLock.Lock()
	defer h.commandLock.Unlock()
	for i := 0; i < 10; i++ { // If the commandID still conflits after trying 10 times, raise error.
		commandID := randomStringFunc(10)
		if _, ok := h.commands[commandID]; ok {
			continue
		}
		h.commands[commandID] = c
		return commandID, nil

	}
	return "", errors.New("can not get a valid commandID")
}

func (h *Handler) StartCommand(ctx context.Context, request *executorpb.StartCommandRequest) (*executorpb.StartCommandResponse, error) {
	rb := lib.NewRingBuffer(defaultRingBufferSize)
	if len(request.Commands) == 0 {
		return nil, errors.Errorf("no command provided")
	}
	log.Infof("starting command on dir: %s", request.Dir)
	cmd := exec.Command(request.Commands[0], request.Commands[1:]...)
	cmd.Dir = request.Dir
	if len(request.Env) > 0 {
		cmd.Env = request.Env
	}
	cmd.Stdout = rb
	cmd.Stderr = rb
	c := newCommand(cmd, rb)
	if err := c.Start(); err != nil {
		return nil, errors.WithMessage(err, "failed to start command")
	}

	commandID, err := h.setCommand(c)
	if err != nil {
		return nil, err
	}

	logger := log.ExtractLogger(ctx)
	go func() {
		c.Wait()
		logger.Infof("process %d exited", c.Process.Pid)
	}()
	return &executorpb.StartCommandResponse{
		CommandID: commandID,
		Status: &executorpb.ProcessStatus{
			Pid:   int32(cmd.Process.Pid),
			Exit:  false,
			Error: "",
		},
	}, nil
}

func (h *Handler) WaitCommand(ctx context.Context, request *executorpb.WaitCommandRequest) (*executorpb.WaitCommandResponse, error) {
	h.commandLock.RLock()
	c, ok := h.commands[request.CommandID]
	h.commandLock.RUnlock()
	if !ok {
		return nil, errors.Errorf("command with pid %s not found", request.CommandID)
	}
	select {
	case <-time.After(time.Duration(request.Timeout)):
		return &executorpb.WaitCommandResponse{
			Status: &executorpb.ProcessStatus{
				Pid:      int32(c.Process.Pid),
				ExitCode: 0,
				Exit:     false,
				Error:    "",
			},
		}, nil
	case <-c.runningCh:
		s := &executorpb.WaitCommandResponse{
			Status: &executorpb.ProcessStatus{
				Pid:      int32(c.Process.Pid),
				ExitCode: int32(c.ProcessState.ExitCode()),
				Exit:     c.ProcessState.Exited(),
			},
		}
		if c.waitError != nil {
			s.Status.Error = c.waitError.Error()
		}
		// Here, we keep the command id.
		// It's a leak, but it's not a big deal. The executor will be ended soon.
		return s, nil

	}
}

type command struct {
	*exec.Cmd
	logWriter io.ReadWriteCloser
	runningCh chan struct{}
	waitError error
}

func newCommand(cmd *exec.Cmd, logWriter io.ReadWriteCloser) *command {
	return &command{
		Cmd:       cmd,
		logWriter: logWriter,
		runningCh: make(chan struct{}),
	}
}

func (c *command) Start() error {
	err := c.Cmd.Start()
	if err != nil {
		close(c.runningCh)
		_ = c.logWriter.Close()
	}
	return err
}

// Wait waits for the command to exit.
func (c *command) Wait() {
	c.waitError = c.Cmd.Wait()
	_ = c.logWriter.Close()
	close(c.runningCh)
}
