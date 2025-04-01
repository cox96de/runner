package starlark

import (
	"io"

	"github.com/cox96de/runner/log"
	"github.com/samber/lo"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

const contextKey = "context"

type context struct {
	stdout   io.Writer
	stderr   io.Writer
	workdir  string
	env      []string
	username string
}

type Command struct {
	context   *context
	logWriter io.ReadWriteCloser
	script    string
	runningCh chan error
	waitError error
}

func (s *Command) Read(buf []byte) (int, error) {
	return s.logWriter.Read(buf)
}

func (s *Command) GetPID() int {
	return 1
}

func (s *Command) ExitCode() int {
	return lo.Ternary(s.waitError == nil, 0, 1)
}

func NewCommand(script string, stdout io.ReadWriteCloser, stderr io.ReadWriteCloser, workDir string, env []string, username string) (*Command, error) {
	return &Command{
		context: &context{
			stdout:   stdout,
			stderr:   stderr,
			workdir:  workDir,
			env:      env,
			username: username,
		},
		logWriter: stdout,
		runningCh: make(chan error),
		script:    script,
	}, nil
}

func (s *Command) run() error {
	thread := &starlark.Thread{
		Print: func(_ *starlark.Thread, msg string) {
			_, err := s.context.stdout.Write([]byte(msg))
			if err != nil {
				log.Errorf("failed to write to stdout: %v", err)
			}
		},
	}
	thread.SetLocal(contextKey, &context{
		stdout:   s.context.stdout,
		stderr:   s.context.stderr,
		workdir:  s.context.workdir,
		env:      s.context.env,
		username: s.context.username,
	})
	functions, err := NewFunctions(thread)
	if err != nil {
		return err
	}
	resultDict, err := starlark.ExecFileOptions(&syntax.FileOptions{
		Set:             true,
		TopLevelControl: true,
	}, thread, "starlark", s.script, functions)
	if err != nil {
		return err
	}
	_ = resultDict
	return nil
}

func (s *Command) Start() error {
	go func() {
		s.waitError = s.run()
		if s.waitError != nil {
			s.runningCh <- s.waitError
		}
		close(s.runningCh)
		_ = s.logWriter.Close()
	}()
	return nil
}

func (s *Command) Wait() <-chan error {
	return s.runningCh
}
