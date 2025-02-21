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

	"github.com/cox96de/runner/app/executor/starlark"

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

var (
	randomStringFunc         = util.RandomString
	_                Command = (*command)(nil)
	_                Command = (*starlark.Command)(nil)
)

type Command interface {
	Read(buf []byte) (int, error)
	GetPID() int
	Wait() <-chan error
	ExitCode() int
	Start() error
}

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
		n, readErr := c.Read(logBuf)
		if n == 0 && readErr != nil {
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

func (h *Handler) setCommand(c Command) (string, error) {
	h.commandLock.Lock()
	defer h.commandLock.Unlock()
	for i := 0; i < 10; i++ { // If the commandID still conflict after trying 10 times, raise error.
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
	if len(request.Commands) == 0 && len(request.Script) == 0 {
		return nil, errors.Errorf("no command or script provided")
	}
	log.Infof("starting command on dir: %s", request.Dir)
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
	var (
		c   Command
		err error
	)
	if len(request.Script) > 0 {
		c, err = starlark.NewCommand(request.Script, rb, rb, request.Dir, request.Env, request.Username)
	} else {
		c, err = newCommand(request.Commands, rb, rb, request.Dir, request.Env, request.Username)
	}
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to create command: %s", request.Commands)
	}
	if err := c.Start(); err != nil {
		return nil, errors.WithMessage(err, "failed to start command")
	}

	commandID, err := h.setCommand(c)
	if err != nil {
		return nil, err
	}
	return &executorpb.StartCommandResponse{
		CommandID: commandID,
		Status: &executorpb.ProcessStatus{
			Pid:   int32(c.GetPID()),
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
	case <-ctx.Done():
		return nil, errors.New("context is done")
	case <-time.After(time.Duration(request.Timeout)):
		return &executorpb.WaitCommandResponse{
			Status: &executorpb.ProcessStatus{
				Pid:      int32(c.GetPID()),
				ExitCode: 0,
				Exit:     false,
				Error:    "",
			},
		}, nil
	case waitError := <-c.Wait():
		s := &executorpb.WaitCommandResponse{
			Status: &executorpb.ProcessStatus{
				Pid:      int32(c.GetPID()),
				ExitCode: int32(c.ExitCode()),
				// Don't use c.ProcessState.Exit().
				// It's false if process is terminated by signal.
				Exit: true,
			},
		}
		if waitError != nil {
			s.Status.Error = waitError.Error()
		}
		// Here, we keep the command id.
		// It's a leak, but it's not a big deal. The executor will be ended soon.
		return s, nil

	}
}

func newCommand(commands []string, stdout io.ReadWriteCloser, stderr io.WriteCloser, workDir string, env []string, username string) (*command, error) {
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = workDir
	_, err := setUser(cmd, username)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to set user: %s", username)
	}
	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}
	return &command{
		Cmd:       cmd,
		logWriter: stdout,
		runningCh: make(chan error),
	}, nil
}

type command struct {
	*exec.Cmd
	logWriter io.ReadWriteCloser
	runningCh chan error
	waitError error
}

func (c *command) Start() error {
	err := c.Cmd.Start()
	if err != nil {
		close(c.runningCh)
		_ = c.logWriter.Close()
		return err
	}
	go func() {
		c.waitError = c.Cmd.Wait()
		_ = c.logWriter.Close()
		if c.waitError != nil {
			c.runningCh <- c.waitError
		}
		close(c.runningCh)
	}()
	return err
}

func (c *command) Read(buf []byte) (int, error) {
	return c.logWriter.Read(buf)
}

// Wait waits for the command to exit.
func (c *command) Wait() <-chan error {
	return c.runningCh
}

func (c *command) GetPID() int {
	return c.Cmd.Process.Pid
}

func (c *command) ExitCode() int {
	return c.ProcessState.ExitCode()
}
