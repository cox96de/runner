package handler

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/lib"
	"github.com/cox96de/runner/log"
	"github.com/cox96de/runner/util"
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

var (
	realUser      *user.User
	effectiveUser *user.User
)

func setUser(cmd *exec.Cmd, username string) (u *user.User, err error) {
	defer func() {
		if u != nil {
			setHomeEnv(cmd, u)
		}
	}()
	if len(username) > 0 {
		if username == effectiveUser.Username {
			return effectiveUser, nil
		}
		user, err := user.Lookup(username)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to find user '%s'", username)
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		return user, setUserForSysProcAttr(cmd.SysProcAttr, user)
	} else if effectiveUser.Uid != realUser.Uid {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		return realUser, setUserForSysProcAttr(cmd.SysProcAttr, realUser)
	}
	return realUser, nil
}

func setHomeEnv(cmd *exec.Cmd, user *user.User) {
	// Only linux needs.
	if runtime.GOOS != "linux" {
		return
	}
	// The executor might be bootstrapped without login shell.
	cmd.Env = append(cmd.Env, "HOME="+user.HomeDir)
}

func (h *Handler) StartCommand(ctx context.Context, request *executorpb.StartCommandRequest) (*executorpb.StartCommandResponse, error) {
	rb := lib.NewRingBuffer(defaultRingBufferSize)
	if len(request.Commands) == 0 {
		return nil, errors.Errorf("no command provided")
	}
	log.Infof("starting command on dir: %s", request.Dir)
	cmd := exec.Command(request.Commands[0], request.Commands[1:]...)
	if len(request.Dir) > 0 {
		stat, err := os.Stat(request.Dir)
		switch {
		case err == nil && !stat.IsDir():
			return nil, errors.Errorf("dir is not a directory: %s", request.Dir)
		case err != nil && os.IsNotExist(err):
			log.Infof("dir is not exist, creating dir: %s", request.Dir)
			if err := mkdirAll(request.Dir, os.ModePerm, request.Username); err != nil {
				return nil, errors.WithMessagef(err, "failed to create dir: %s", request.Dir)
			}
		case err != nil:
			return nil, errors.WithMessagef(err, "failed to stat dir: %s", request.Dir)
		}
	}
	cmd.Dir = request.Dir
	_, err := setUser(cmd, request.Username)
	if err != nil {
		return nil, err
	}
	if len(request.Env) > 0 {
		cmd.Env = append(cmd.Env, request.Env...)
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
				// Don't use c.ProcessState.Exit().
				// It's false if process is terminated by signal.
				Exit: true,
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
