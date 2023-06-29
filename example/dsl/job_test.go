package dsl

import (
	"testing"

	"github.com/cox96de/runner/testtool"

	"github.com/cox96de/runner/engine"
	"gopkg.in/yaml.v2"
	"gotest.tools/v3/assert"
)

func TestParseDSL(t *testing.T) {
	j := &Job{
		Runner: &Runner{Kube: &engine.KubeSpec{
			Containers: []*engine.Container{{
				Name:  "test",
				Image: "debian",
			}},
		}},
		DefaultContainerName: "test",
		Steps:                []*Step{{Commands: []string{"echo hello"}}},
	}
	out, err := yaml.Marshal(j)
	if err != nil {
		panic(err)
	}
	nj, err := ParseDSL(out)
	assert.NilError(t, err)
	testtool.DeepEqualObject(t, nj, "testdata/parse_dsl.json")
}
