package handler

import (
	"context"
	"testing"

	"github.com/cox96de/runner/app/server/eventhook"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/mock"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

func TestHandler_UpdateStepExecution(t *testing.T) {
	handler := NewHandler(mock.NewMockDB(t), nil, nil, nil, nil, eventhook.NewService())
	executions, err := handler.db.CreateStepExecutions(context.Background(), []*db.CreateStepExecutionOption{
		{
			JobExecutionID: 1,
			StepID:         1,
			Status:         api.StatusCreated,
		},
	})
	assert.NilError(t, err)
	updateStepExecutionResponse, err := handler.UpdateStepExecution(context.Background(), &api.UpdateStepExecutionRequest{
		StepExecutionID: executions[0].ID,
		Status:          lo.ToPtr(api.StatusRunning),
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, updateStepExecutionResponse.StepExecution.Status, api.StatusRunning)
	t.Run("GetStepExecution", func(t *testing.T) {
		getStepExecutionResponse, err := handler.GetStepExecution(context.Background(), &api.GetStepExecutionRequest{
			StepExecutionID: executions[0].ID,
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, getStepExecutionResponse.StepExecution.Status, api.StatusRunning)
	})
}
