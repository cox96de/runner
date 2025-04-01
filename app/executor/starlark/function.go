package starlark

import (
	_ "embed"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cockroachdb/errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

//go:embed function.py
var function string

func runtimeGOOS(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.String(runtime.GOOS), nil
}

func osEnvironment(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c := thread.Local(contextKey).(*context)
	envs := starlark.NewDict(len(c.env))
	for _, e := range c.env {
		sp := strings.SplitN(e, "=", 2)
		err := envs.SetKey(starlark.String(sp[0]), starlark.String(sp[1]))
		if err != nil {
			return nil, err
		}
	}
	return envs, nil
}

func commandRun(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		arg starlark.Value
		cwd string
		env *starlark.Dict
	)
	c := thread.Local(contextKey).(*context)
	err := starlark.UnpackArgs(fn.Name(), args, kwargs, "args", &arg, "cwd?", &cwd, "env?", &env)
	if err != nil {
		return nil, errors.WithMessage(err, "parse args")
	}
	list := arg.(*starlark.List)
	var argsList []string
	for i := 0; i < list.Len(); i++ {
		s, _ := starlark.AsString(list.Index(i))
		argsList = append(argsList, s)
	}
	var envs []string
	for _, tuple := range env.Items() {
		envs = append(envs, fmt.Sprintf("%s=%s", tuple[0].String(), tuple[1].String()))
	}
	command := exec.Command(argsList[0], argsList[1:]...)
	command.Dir = c.workdir
	command.Env = envs
	command.Stdout = c.stdout
	command.Stderr = c.stderr
	return starlark.None, command.Run()
}

func NewFunctions(thread *starlark.Thread) (starlark.StringDict, error) {
	dict := starlark.StringDict{
		"struct":         starlark.NewBuiltin("struct", starlarkstruct.Make),
		"_commandRun":    starlark.NewBuiltin("_commandRun", commandRun),
		"_runtimeGOOS":   starlark.NewBuiltin("_runtimeGOOS", runtimeGOOS),
		"_osEnvironment": starlark.NewBuiltin("_osEnvironment", osEnvironment),
	}
	resultDict, err := starlark.ExecFileOptions(&syntax.FileOptions{
		Set:             true,
		TopLevelControl: true,
	}, thread, "function", function, dict)
	return resultDict, err
}
