package db

import (
	"context"
	"testing"
	"time"

	"github.com/cox96de/runner/api"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"

	"gotest.tools/v3/assert"
)

func TestClient_CreateJobs(t *testing.T) {
	db := NewMockDB(t, &Job{})
	jobs, err := db.CreateJobs(context.Background(), []*CreateJobOption{
		{
			PipelineID: 1,
			Name:       "job1",
			RunsOn: &api.RunsOn{
				Label: "label_1",
			},
			WorkingDirectory: "/home/user",
			EnvVar:           map[string]string{"env1": "val1"},
			DependsOn:        []string{"job2"},
		},
		{
			PipelineID: 1,
			Name:       "job2",
			RunsOn: &api.RunsOn{
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
	jobExecutions, err := db.CreateJobExecutions(context.Background(), []*CreateJobExecutionOption{
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
	assert.DeepEqual(t, jobExecutions, []*JobExecution{
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
		for _, job := range jobExecutions {
			jobByID, err := db.GetJobExecution(context.Background(), job.ID)
			assert.NilError(t, err, job.JobID)
			assert.DeepEqual(t, job, jobByID)
		}
	})
	t.Run("GetJobExecutionsByJobID", func(t *testing.T) {
		jobExecutions, err := db.GetJobExecutionsByJobID(context.Background(), 1)
		assert.NilError(t, err)
		assert.DeepEqual(t, jobExecutions, []*JobExecution{jobExecutions[0]})
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
	updatedJob, err := db.UpdateJobExecution(context.Background(), &UpdateJobExecutionOption{
		ID:     executions[0].ID,
		Status: lo.ToPtr(api.StatusRunning),
	})
	assert.NilError(t, err)
	assert.Equal(t, api.StatusRunning, updatedJob.Status)
	execution, err := db.GetJobExecution(context.Background(), executions[0].ID)
	assert.NilError(t, err)
	assert.Equal(t, api.StatusRunning, execution.Status)
}

func TestClient_ResetJobExecution(t *testing.T) {
	db := NewMockDB(t, &JobExecution{})
	executions, err := db.CreateJobExecutions(context.Background(), []*CreateJobExecutionOption{
		{
			JobID:  1,
			Status: api.StatusCreated,
		},
	})
	assert.NilError(t, err)
	_, err = db.UpdateJobExecution(context.Background(), &UpdateJobExecutionOption{
		ID:     executions[0].ID,
		Status: lo.ToPtr(api.StatusFailed),
		Reason: &api.Reason{
			Reason:  api.FailedReasonStepFailed,
			Message: "step failed",
		},
		StartedAt:   lo.ToPtr(time.Now()),
		CompletedAt: lo.ToPtr(time.Now()),
	})
	assert.NilError(t, err)
	err = db.ResetJobExecution(context.Background(), executions[0].ID)
	assert.NilError(t, err)
	execution, err := db.GetJobExecution(context.Background(), executions[0].ID)
	assert.NilError(t, err)
	assert.Equal(t, api.StatusCreated, execution.Status)
	assert.Equal(t, "", string(execution.Reason))
	assert.Assert(t, execution.StartedAt == nil)
	assert.Assert(t, execution.CompletedAt == nil)
}
