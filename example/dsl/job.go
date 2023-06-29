package dsl

import (
	"github.com/cox96de/runner/engine"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Job is the job to run.
type Job struct {
	// Runner is the runner spec. It describes the environment of the job to run.
	Runner *Runner
	// DefaultContainerName is the default container name to run the commands.
	// It's only used when the step doesn't specify the container name.
	// For container based runner, it is the container name.
	DefaultContainerName string
	// Steps is the steps to run.
	// All steps will be run in order and in the same environment (which defined in Runner).
	Steps []*Step
}

// Runner is the runner spec. It describes the environment of the job to run.
type Runner struct {
	// Kube is the spec for kubernetes based runner.
	// Job will be run in a kubernetes pod.
	Kube *engine.KubeSpec
}
type Step struct {
	// Workdir is the working directory of the step.
	// It should be an absolute path.
	Workdir string
	// Commands is the commands to run.
	Commands []string
	// ContainerName is the container name to run the commands.
	// If it's empty, the DefaultContainerName of the job will be used.
	ContainerName string
}

// ParseDSL parses the DSL content to a job.
func ParseDSL(content []byte) (*Job, error) {
	j := &Job{}
	err := yaml.Unmarshal(content, j)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return j, nil
}
