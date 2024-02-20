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
	t.Run("GetStepsByJobID", func(t *testing.T) {
		steps, err := db.GetStepsByJobID(context.Background(), 1)
		assert.NilError(t, err)
		assert.DeepEqual(t, steps, steps, cmpopts.IgnoreFields(Step{}, "ID", "CreatedAt", "UpdatedAt"))
	})
}

func TestClient_CreateStepExecutions(t *testing.T) {
	db := NewMockDB(t, &StepExecution{})
	stepExecutions, err := db.CreateStepExecutions(context.Background(), []*CreateStepExecutionOption{
		{
			JobExecutionID: 1,
			StepID:         1,
			Status:         0,
		},
		{
			JobExecutionID: 1,
			StepID:         2,
			Status:         0,
		},
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, stepExecutions, []*StepExecution{
		{
			JobExecutionID: 1,
			StepID:         1,
			Status:         0,
		},
		{
			JobExecutionID: 1,
			StepID:         2,
			Status:         0,
		},
	}, cmpopts.IgnoreFields(StepExecution{}, "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"))
	t.Run("GetStepExecutionsByJobExecutionID", func(t *testing.T) {
		stepExecutions, err := db.GetStepExecutionsByJobExecutionID(context.Background(), 1)
		assert.NilError(t, err)
		assert.DeepEqual(t, stepExecutions, stepExecutions, cmpopts.IgnoreFields(StepExecution{}, "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"))
	})
}
