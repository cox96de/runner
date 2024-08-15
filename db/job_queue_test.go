package db

import (
	"context"
	"testing"
	"time"

	"github.com/cox96de/runner/api"
	"gotest.tools/v3/assert"
)

func TestClient_GetQueuedJobExecutionIDs(t *testing.T) {
	db := NewMockDB(t, &JobQueue{})
	t.Run("empty", func(t *testing.T) {
		executions, err := db.GetQueuedJobExecutionIDs(context.Background(), "label", 10)
		assert.NilError(t, err)
		assert.Assert(t, len(executions) == 0)
	})
	_, err := db.CreateJobQueues(context.Background(), []*CreateJobQueueOption{
		{
			JobExecutionID: 1,
			Status:         api.StatusQueued,
			Label:          "label1",
		},
		{
			JobExecutionID: 2,
			Status:         api.StatusQueued,
			Label:          "label1",
		},
		{
			JobExecutionID: 3,
			Status:         api.StatusRunning,
			Label:          "label2",
		},
		{
			JobExecutionID: 4,
			Status:         api.StatusQueued,
			Label:          "label2",
		},
	})
	t.Run("with_match_label", func(t *testing.T) {
		assert.NilError(t, err)
		executions, err := db.GetQueuedJobExecutionIDs(context.Background(), "label1", 1)
		assert.NilError(t, err)
		assert.Assert(t, len(executions) == 1)
		assert.Equal(t, api.StatusQueued, executions[0].Status)
		jobExecutions, err := db.GetQueuedJobExecutionIDs(context.Background(), "label1", 10)
		assert.NilError(t, err)
		assert.Assert(t, len(jobExecutions) == 2)
	})
	t.Run("with_match_label2", func(t *testing.T) {
		assert.NilError(t, err)
		executions, err := db.GetQueuedJobExecutionIDs(context.Background(), "label2", 10)
		assert.NilError(t, err)
		assert.Assert(t, len(executions) == 1)
	})
}

func TestClient_ListHeartbeatJobExecutions(t *testing.T) {
	db := NewMockDB(t, &JobQueue{})
	jobQueues, err := db.CreateJobQueues(context.Background(), []*CreateJobQueueOption{{
		JobExecutionID: 1,
		Label:          "",
		Status:         0,
	}})
	assert.NilError(t, err)
	err = db.TouchHeartbeat(context.Background(), jobQueues[0].JobExecutionID)
	assert.NilError(t, err)
	executions, err := db.ListHeartbeatJobExecutions(context.Background(), time.Second)
	assert.NilError(t, err)
	assert.Assert(t, len(executions) == 0)
	time.Sleep(time.Microsecond * 10)
	executions, err = db.ListHeartbeatJobExecutions(context.Background(), 0)
	assert.NilError(t, err)
	assert.Assert(t, len(executions) == 1)
	assert.Assert(t, executions[0].JobExecutionID == jobQueues[0].JobExecutionID)
}
