package main

import (
	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/engine/shell"
	"github.com/pkg/errors"
)

func ComposeEngine(c *Config) (engine.Engine, error) {
	switch c.Engine.Name {
	case "shell":
		return shell.NewEngine(), nil
	default:
		return nil, errors.Errorf("unsupported engine '%s'", c.Engine.Name)
	}
}
