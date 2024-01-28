package shell

import (
	"context"

	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/entity"
)

type Engine struct{}

func (e *Engine) Ping(_ context.Context) error {
	return nil
}

func (e *Engine) CreateRunner(_ context.Context, _ *entity.Job) (engine.Runner, error) {
	return NewRunner(), nil
}
