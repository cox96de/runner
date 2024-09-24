package dag

import (
	"github.com/cockroachdb/errors"
	"github.com/natessilva/dag"
)

// Runner executes functions as DAG.
type Runner struct {
	*dag.Runner
}

// NewRunner creates a DAG runner.
func NewRunner() *Runner {
	return &Runner{
		Runner: &dag.Runner{},
	}
}

// AddVertex adds a function as a vertex in the graph.
func (r *Runner) AddVertex(name string, fn func() error) {
	r.Runner.AddVertex(name, func() (recoveredErr error) {
		// Handle panic.
		defer func() {
			if err := recover(); err != nil {
				recoveredErr = errors.Errorf("recover from vertex: %+v", err)
			}
		}()
		return fn()
	})
}
