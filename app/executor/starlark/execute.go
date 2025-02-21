package starlark

import (
	"io"
	"os/exec"
	"runtime"

	"github.com/cox96de/runner/log"
	"github.com/samber/lo"

	"github.com/cockroachdb/errors"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func platformSystem(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	switch runtime.GOOS {
	case "windows":
		return starlark.String("Windows"), nil
	case "darwin":
		return starlark.String("Darwin"), nil
	case "linux":
		return starlark.String("Linux"), nil
	default:
		return starlark.String("Unknown"), nil
	}
}

const contextKey = "context"

type context struct {
	stdout   io.Writer
	stderr   io.Writer
	workdir  string
	env      []string
	username string
}

func subprocessRun(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		arg   starlark.Value
		shell                = false
		cwd   starlark.Value = starlark.None
	)
	c := thread.Local(contextKey).(*context)
	err := starlark.UnpackArgs(fn.Name(), args, kwargs, "args", &arg, "shell?", &shell, "cwd?", &cwd)
	if err != nil {
		return nil, errors.WithMessage(err, "parse args")
	}
	if args.Len() != 1 {
		return starlark.None, errors.New("expected 1 argument")
	}
	list := arg.(*starlark.List)
	var argsList []string
	for i := 0; i < list.Len(); i++ {
		s, _ := starlark.AsString(list.Index(i))
		argsList = append(argsList, s)
	}
	command := exec.Command(argsList[0], argsList[1:]...)
	command.Dir = c.workdir
	command.Env = c.env
	command.Stdout = c.stdout
	command.Stderr = c.stderr
	return starlark.None, command.Run()
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
	dict := starlark.StringDict{
		"platform_system": starlark.NewBuiltin("platform_system", platformSystem),
		"process_run":     starlark.NewBuiltin("process_run", subprocessRun),
	}
	resultDict, err := starlark.ExecFileOptions(&syntax.FileOptions{
		Set:             true,
		TopLevelControl: true,
	}, thread, "starlark", s.script, dict)
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
