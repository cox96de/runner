package api

import (
	"testing"

	"gotest.tools/v3/assert"
)

func getDSL() *PipelineDSL {
	return &PipelineDSL{
		unknownFields: nil,
		Jobs: []*JobDSL{
			{
				Name: "uid",
				RunsOn: &RunsOn{
					Label: "label",
				},
				Steps: []*StepDSL{
					{
						Name:     "step_uid",
						Commands: []string{"echo 'hello'"},
					},
				},
			},
		},
	}
}

func TestValidateDSL(t *testing.T) {
	dsl := getDSL()
	err := ValidateDSL(dsl)
	assert.NilError(t, err)

	dsl = getDSL()
	dsl.Jobs = nil
	assert.Assert(t, ValidateDSL(dsl) != nil)

	dsl = getDSL()
	dsl.Jobs[0].RunsOn = nil
	assert.Assert(t, ValidateDSL(dsl) != nil)

	dsl = getDSL()
	dsl.Jobs[0].RunsOn.Label = ""
	assert.Assert(t, ValidateDSL(dsl) != nil)

	dsl = getDSL()
	dsl.Jobs[0].Name = ""
	assert.Assert(t, ValidateDSL(dsl) != nil)
}
