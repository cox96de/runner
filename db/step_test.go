package db

import (
	"context"
	"testing"
	"time"

	"github.com/cox96de/runner/api"
	"github.com/samber/lo"

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
		got, err := db.GetStepExecutionsByJobExecutionID(context.Background(), 1)
		assert.NilError(t, err)
		assert.DeepEqual(t, got, stepExecutions, cmpopts.IgnoreFields(StepExecution{}, "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"))
	})
	t.Run("GetStepExecutionsByJobExecutionIDs", func(t *testing.T) {
		got, err := db.GetStepExecutionsByJobExecutionIDs(context.Background(), []int64{1})
		assert.NilError(t, err)
		assert.DeepEqual(t, got, map[int64][]*StepExecution{1: stepExecutions}, cmpopts.IgnoreFields(StepExecution{}, "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"))
	})
	t.Run("GetStepExecution", func(t *testing.T) {
		for _, stepExecution := range stepExecutions {
			stepExecutionByID, err := db.GetStepExecution(context.Background(), stepExecution.ID)
			assert.NilError(t, err)
			assert.DeepEqual(t, stepExecution, stepExecutionByID)
		}
	})
}

func TestClient_UpdateStepExecution(t *testing.T) {
	db := NewMockDB(t, &StepExecution{})
	stepExecutions, err := db.CreateStepExecutions(context.Background(), []*CreateStepExecutionOption{
		{
			JobExecutionID: 1,
			StepID:         1,
			Status:         0,
		},
	})
	assert.NilError(t, err)
	execution := stepExecutions[0]
	_, err = db.UpdateStepExecution(context.Background(), &UpdateStepExecutionOption{
		ID:     execution.ID,
		Status: lo.ToPtr(api.StatusRunning),
	})
	assert.NilError(t, err)
	execution, err = db.GetStepExecution(context.Background(), execution.ID)
	assert.NilError(t, err)
	assert.Equal(t, api.StatusRunning, execution.Status)
	updatedStep, err := db.UpdateStepExecution(context.Background(), &UpdateStepExecutionOption{
		ID:       execution.ID,
		ExitCode: lo.ToPtr(uint32(1)),
	})
	assert.NilError(t, err)
	assert.Equal(t, updatedStep.ExitCode, uint32(1))
	execution, err = db.GetStepExecution(context.Background(), execution.ID)
	assert.NilError(t, err)
	assert.Equal(t, uint32(1), execution.ExitCode)
}

func TestClient_ResetStepExecutions(t *testing.T) {
	db := NewMockDB(t, &StepExecution{})
	executions, err := db.CreateStepExecutions(context.Background(), []*CreateStepExecutionOption{
		{
			JobExecutionID: 1,
			StepID:         2,
			Status:         api.StatusCreated,
		},
	})
	assert.NilError(t, err)
	_, err = db.UpdateStepExecution(context.Background(), &UpdateStepExecutionOption{
		ID:          executions[0].ID,
		Status:      lo.ToPtr(api.StatusFailed),
		ExitCode:    lo.ToPtr(uint32(1)),
		StartedAt:   lo.ToPtr(time.Now()),
		CompletedAt: lo.ToPtr(time.Now()),
	})
	assert.NilError(t, err)
	err = db.ResetStepExecutions(context.Background(), []int64{executions[0].ID})
	assert.NilError(t, err)
	execution, err := db.GetStepExecution(context.Background(), executions[0].ID)
	assert.NilError(t, err)
	assert.Equal(t, api.StatusCreated, execution.Status)
	assert.Equal(t, uint32(0), execution.ExitCode)
	assert.Equal(t, (*time.Time)(nil), execution.StartedAt)
	assert.Equal(t, (*time.Time)(nil), execution.CompletedAt)
}
