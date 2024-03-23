package shell

import (
	"context"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/engine"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Ping(_ context.Context) error {
	return nil
}

func (e *Engine) CreateRunner(_ context.Context, _ engine.LogProvider, _ *api.Job) (engine.Runner, error) {
	return NewRunner(), nil
}
