package dispatch

import (
	"context"
	"testing"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/mock"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

func TestUpdateJobExecution(t *testing.T) {
	dbClient := mock.NewMockDB(t)
	service := NewService(dbClient)
	jobs, err := dbClient.CreateJobs(context.Background(), []*db.CreateJobOption{
		{
			PipelineID: 1,
			Name:       t.Name(),
		},
	})
	assert.NilError(t, err)
	job := jobs[0]
	executions, err := dbClient.CreateJobExecutions(context.Background(), []*db.CreateJobExecutionOption{
		{
			JobID:  job.ID,
			Status: api.StatusCreated,
		},
	})
	assert.NilError(t, err)
	execution := executions[0]
	t.Run("to_queuing", func(t *testing.T) {
		err = service.UpdateJobExecution(context.Background(), dbClient, &db.UpdateJobExecutionOption{
			ID:     execution.ID,
			Status: lo.ToPtr(api.StatusQueued),
		})
		assert.NilError(t, err)
		jobQueue, err := dbClient.GetJobQueue(context.Background(), execution.ID)
		assert.NilError(t, err)
		assert.Equal(t, jobQueue.Status, api.StatusQueued)
		t.Run("to_running", func(t *testing.T) {
			err = service.UpdateJobExecution(context.Background(), dbClient, &db.UpdateJobExecutionOption{
				ID:     execution.ID,
				Status: lo.ToPtr(api.StatusRunning),
			})
			assert.NilError(t, err)
			jobQueue, err := dbClient.GetJobQueue(context.Background(), execution.ID)
			assert.NilError(t, err)
			assert.Equal(t, jobQueue.Status, api.StatusRunning)
		})
		t.Run("to_completed", func(t *testing.T) {
			err = service.UpdateJobExecution(context.Background(), dbClient, &db.UpdateJobExecutionOption{
				ID:     execution.ID,
				Status: lo.ToPtr(api.StatusFailed),
			})
			assert.NilError(t, err)
			_, err := dbClient.GetJobQueue(context.Background(), execution.ID)
			assert.Assert(t, db.IsRecordNotFoundError(err))
		})
	})
}
