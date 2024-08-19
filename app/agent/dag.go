package agent

import (
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/lib"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// dagNode wrap api.Step to implement dag node.
type dagNode struct {
	// depends override api.Step.DependsOn.
	// If depends is not empty, it will be used as depends.
	depends []string
	*api.Step
}

func (d *dagNode) ID() string {
	return d.Name
}

func (d *dagNode) Depends() []string {
	if d.depends != nil {
		return d.depends
	}
	return d.DependsOn
}

func isDAGStep(steps []*api.Step) bool {
	for _, step := range steps {
		if len(step.DependsOn) > 0 {
			return true
		}
	}
	return false
}

func (e *Execution) normalizeDAG() error {
	dagStep := isDAGStep(e.job.Steps)
	dagNodes := make([]*dagNode, 0, len(e.job.Steps))
	if dagStep {
		dagNodes = lo.Map(e.job.Steps, func(step *api.Step, index int) *dagNode {
			return &dagNode{
				Step: step,
			}
		})
	} else {
		preStepName := e.job.Steps[0].Name
		dagNodes = append(dagNodes, &dagNode{
			Step: e.job.Steps[0],
		})
		for i := 1; i < len(e.job.Steps); i++ {
			dagNodes = append(dagNodes, &dagNode{
				Step:    e.job.Steps[i],
				depends: []string{preStepName},
			})
			preStepName = e.job.Steps[i].Name
		}
	}
	dag, err := lib.NewDAG[*dagNode](dagNodes...)
	if err != nil {
		return errors.WithMessage(err, "failed to calculate DAG")
	}
	e.dag = dag
	return err
}
