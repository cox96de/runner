package agent

import (
	"context"
	"github.com/cox96de/runner/api"
	"github.com/pkg/errors"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type Context struct {
	Job  *api.Job  `json:"job"`
	Step *api.Step `json:"step"`
}

// Context is the context of the execution.
//
//	context
//	  |- job
//	     |- execution
//	     |- steps
//	  |  step
//	  |  step
func (e *Execution) continueWhenNoPreFailed(step *api.Step) (bool, error) {
	deepPres, err := e.dag.DeepPre(step.Name)
	if err != nil {
		return false, errors.WithMessage(err, "") //TODO
	}
	for _, pre := range deepPres {
		if e.stepExecutions[pre.Step.ID].Status == api.StatusFailed {
			return false, nil
		}
	}
	return true, nil
}

func (e *Execution) preStep(step *api.Step) (bool, error) {
	if step.If == "" {
		return e.continueWhenNoPreFailed(step)
	}
	return false, nil
}
func (e *Execution) onPreviousSuccess(ctx context.Context, c *Context) {

}
func (e *Execution) evalExpression(ctx context.Context, expression string) (bool, error) {
	thread := &starlark.Thread{
		Name: "main",
		Print: func(thread *starlark.Thread, msg string) {
			// Nop print.
		},
	}
	stringDict := starlark.StringDict{
		"true":  starlark.True,
		"false": starlark.False,
	}
	options, err := starlark.EvalOptions(&syntax.FileOptions{}, thread, "expression", expression,
		stringDict)
	if err != nil {
		return false, errors.WithMessage(err, "")
	}
	switch options := options.(type) {
	case starlark.Bool:
		return bool(options), nil
	case starlark.String:
		return string(options) == "True" || string(options) == "true", nil
	default:
		return false, errors.Errorf("unsupported expression type, expect bool or string type, got %s", options.Type())
	}
}
