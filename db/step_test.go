package db

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	"gotest.tools/v3/assert"
)

func TestClient_CreateSteps(t *testing.T) {
	db := NewMockDB(t, &Step{})
	steps, err := db.CreateSteps(context.Background(), []*CreateStepOption{
		{
			PipelineID: 1,
			JobID:      1,
			Name:       "step1",
			User:       "root",
		},
	})
	assert.NilError(t, err)
	for _, step := range steps {
		assert.Assert(t, step.ID > 0, step.Name)
		stepByID, err := db.GetStepByID(context.Background(), step.ID)
		assert.NilError(t, err, step.Name)
		assert.DeepEqual(t, step, stepByID)
	}
}

func TestClient_CreateStepExecutions(t *testing.T) {
	db := NewMockDB(t, &StepExecution{})
	jobs, err := db.CreateStepExecutions(context.Background(), []*CreateStepExecutionOption{
		{
			JobExecutionID: 1,
			StepID:         1,
			Status:         0,
		},
		{
			JobExecutionID: 2,
			StepID:         2,
			Status:         0,
		},
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, jobs, []*StepExecution{
		{
			JobExecutionID: 1,
			StepID:         1,
			Status:         0,
		},
		{
			JobExecutionID: 2,
			StepID:         2,
			Status:         0,
		},
	}, cmpopts.IgnoreFields(StepExecution{}, "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"))
}
