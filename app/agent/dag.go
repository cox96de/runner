package agent

import "github.com/cox96de/runner/api"

// dagNode wrap api.Step to implement dag node.
type dagNode struct {
	*api.Step
}

func (d *dagNode) ID() string {
	return d.Name
}

func (d *dagNode) Depends() []string {
	return d.DependsOn
}
