package dsl

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseFromContent(t *testing.T) {
	content := `
jobs:
  job1:
     name: "job_name"
     runs-on: 
       container-image: "debian"
     steps:
       - name: "step1"
         run: 
           - "echo hello"`
	p, err := ParseFromContent([]byte(content))
	assert.NilError(t, err)
	assert.DeepEqual(t, p, &Pipeline{Jobs: map[string]*Job{"job1": {
		JobID: "job1",
		Name:  "job_name",
		RunsOn: &RunsOn{
			ContainerImage: "debian",
		},
		Steps: []*Step{
			{
				Name: "step1",
				Run:  []string{"echo hello"},
			},
		},
	}}})
}
