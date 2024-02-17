package db

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"

	"github.com/cox96de/runner/entity"
	"gotest.tools/v3/assert"
)

func TestClient_CreateJobs(t *testing.T) {
	db := NewMockDB(t, &Job{})
	jobs, err := db.CreateJobs(context.Background(), []*CreateJobOption{
		{
			PipelineID: 1,
			Name:       "job1",
			RunsOn: &entity.RunsOn{
				Label: "label_1",
			},
			WorkingDirectory: "/home/user",
			EnvVar:           map[string]string{"env1": "val1"},
			DependsOn:        []string{"job2"},
		},
		{
			PipelineID: 1,
			Name:       "job2",
			RunsOn: &entity.RunsOn{
				Label: "label_1",
			},
			WorkingDirectory: "/home/user",
			EnvVar:           map[string]string{"env2": "val2"},
		},
	})
	assert.NilError(t, err)
	for _, job := range jobs {
		assert.Assert(t, job.ID > 0, job.Name)
		jobByID, err := db.GetJobByID(context.Background(), job.ID)
		assert.NilError(t, err, job.Name)
		assert.DeepEqual(t, job, jobByID)
	}
}

func TestClient_CreateJobExecutions(t *testing.T) {
	db := NewMockDB(t, &JobExecution{})
	jobs, err := db.CreateJobExecutions(context.Background(), []*CreateJobExecutionOption{
		{
			JobID:  1,
			Status: 0,
		},
		{
			JobID:  2,
			Status: 0,
		},
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, jobs, []*JobExecution{
		{
			JobID:  1,
			Status: 0,
		},
		{
			JobID:  2,
			Status: 0,
		},
	}, cmpopts.IgnoreFields(JobExecution{}, "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"))
	t.Run("GetJobByID", func(t *testing.T) {
		for _, job := range jobs {
			jobByID, err := db.GetJobExecution(context.Background(), job.ID)
			assert.NilError(t, err, job.JobID)
			assert.DeepEqual(t, job, jobByID)
		}
	})
}

func TestClient_UpdateJobExecution(t *testing.T) {
	db := NewMockDB(t, &JobExecution{})
	executions, err := db.CreateJobExecutions(context.Background(), []*CreateJobExecutionOption{
		{
			JobID: 1,
		},
	})
	assert.NilError(t, err)
	err = db.UpdateJobExecution(context.Background(), &UpdateJobExecutionOption{
		ID:     executions[0].ID,
		Status: lo.ToPtr(entity.JobStatusRunning),
	})
	assert.NilError(t, err)
	execution, err := db.GetJobExecution(context.Background(), executions[0].ID)
	assert.NilError(t, err)
	assert.Equal(t, entity.JobStatusRunning, execution.Status)
}

func TestClient_GetQueuedJobExecutions(t *testing.T) {
	db := NewMockDB(t, &JobExecution{})
	t.Run("empty", func(t *testing.T) {
		executions, err := db.GetQueuedJobExecutions(context.Background(), 1)
		assert.NilError(t, err)
		assert.Assert(t, len(executions) == 0)
	})
	_, err := db.CreateJobExecutions(context.Background(), []*CreateJobExecutionOption{
		{
			JobID:  1,
			Status: entity.JobStatusQueued,
		},
		{
			JobID:  2,
			Status: entity.JobStatusQueued,
		},
		{
			JobID:  2,
			Status: entity.JobStatusRunning,
		},
	})
	assert.NilError(t, err)
	executions, err := db.GetQueuedJobExecutions(context.Background(), 1)
	assert.NilError(t, err)
	assert.Assert(t, len(executions) == 1)
	assert.Equal(t, entity.JobStatusQueued, executions[0].Status)
	jobExecutions, err := db.GetQueuedJobExecutions(context.Background(), 100)
	assert.NilError(t, err)
	assert.Assert(t, len(jobExecutions) == 2)
}
