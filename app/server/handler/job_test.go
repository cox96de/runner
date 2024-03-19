package handler

import (
	"context"
	"testing"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/mock"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

func TestHandler_UpdateJobExecution(t *testing.T) {
	handler := NewHandler(mock.NewMockDB(t), nil, nil, mock.NewMockLocker(), nil)
	executions, err := handler.db.CreateJobExecutions(context.Background(), []*db.CreateJobExecutionOption{
		{
			JobID:  1,
			Status: api.StatusCreated,
		},
	})
	assert.NilError(t, err)
	t.Run("bad_status_transmit", func(t *testing.T) {
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobID:          1,
			JobExecutionID: executions[0].ID,
			Status:         lo.ToPtr(api.StatusPreparing),
		})
		assert.Assert(t, err != nil)
	})
	t.Run("transmit", func(t *testing.T) {
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobID:          1,
			JobExecutionID: executions[0].ID,
			Status:         lo.ToPtr(api.StatusQueued),
		})
		assert.NilError(t, err)
		execution, err := handler.db.GetJobExecution(context.Background(), executions[0].ID)
		assert.NilError(t, err)
		assert.DeepEqual(t, execution.Status, api.StatusQueued)
	})
	t.Run("ListJobExecutions", func(t *testing.T) {
		listJobExecutionsResponse, err := handler.ListJobExecutions(context.Background(), &api.ListJobExecutionsRequest{
			JobID: 1,
		})
		assert.NilError(t, err)
		assert.Assert(t, len(listJobExecutionsResponse.Jobs) == 1)
		assert.Assert(t, listJobExecutionsResponse.Jobs[0].ID == executions[0].ID)
	})
}
