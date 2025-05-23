package dsl

import "gopkg.in/yaml.v3"

type Pipeline struct {
	Jobs map[string]*Job `yaml:"jobs"`
}

type Job struct {
	JobID  string  `yaml:"-"`
	Name   string  `yaml:"name"`
	RunsOn *RunsOn `yaml:"runs-on"`
	Steps  []*Step `yaml:"steps"`
}

type RunsOn struct {
	ContainerImage string `yaml:"container-image"`
	Image          string `yaml:"image"`
}

type Step struct {
	Name   string   `yaml:"name"`
	Run    []string `yaml:"run"`
	Script string   `yaml:"-"`
	// Don't open this feature now.
	Env map[string]string `yaml:"-"`
}

// ParseFromContent parse yaml content into Pipeline dsl.
func ParseFromContent(content []byte) (*Pipeline, error) {
	y := &Pipeline{}
	err := yaml.Unmarshal(content, y)
	if err != nil {
		return nil, err
	}
	for jobID, job := range y.Jobs {
		job.JobID = jobID
	}
	return y, nil
}
