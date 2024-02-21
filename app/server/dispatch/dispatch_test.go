package dispatch

import (
	"context"
	"testing"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
	"github.com/samber/lo"

	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/mock"
	"gotest.tools/v3/assert"
)

func TestService_Dispatch(t *testing.T) {
	dbClient := mock.NewMockDB(t)
	pipelineService := pipeline.NewService(dbClient)
	t.Run("no_dep", func(t *testing.T) {
		createdPipeline, err := pipelineService.CreatePipeline(context.Background(), &api.Pipeline{
			Jobs: []*api.Job{
				{
					Name:  "job1",
					Steps: []*api.Step{{Name: "step1"}},
				},
				{
					Name:  "job2",
					Steps: []*api.Step{{Name: "step1"}},
				},
			},
		})
		assert.NilError(t, err)
		s := NewService(dbClient)
		err = s.Dispatch(context.Background(), createdPipeline.CreatedJobs, createdPipeline.CreatedJobExecutions)
		assert.NilError(t, err)
		for _, jobExecution := range createdPipeline.CreatedJobExecutions {
			execution, err := dbClient.GetJobExecution(context.Background(), jobExecution.ID)
			assert.NilError(t, err)
			assert.Equal(t, api.StatusQueued, execution.Status)
		}
	})
	t.Run("dep", func(t *testing.T) {
		createdPipeline, err := pipelineService.CreatePipeline(context.Background(), &api.Pipeline{
			Jobs: []*api.Job{
				{
					Name:  "job1",
					Steps: []*api.Step{{Name: "step1"}},
				},
				{
					Name:      "job2",
					Steps:     []*api.Step{{Name: "step1"}},
					DependsOn: []string{"job1"},
				},
			},
		})
		assert.NilError(t, err)
		s := NewService(dbClient)
		err = s.Dispatch(context.Background(), createdPipeline.CreatedJobs, createdPipeline.CreatedJobExecutions)
		assert.NilError(t, err)
		jobIDNameMap := lo.SliceToMap(createdPipeline.CreatedJobs, func(item *db.Job) (int64, string) {
			return item.ID, item.Name
		})
		for _, jobExecution := range createdPipeline.CreatedJobExecutions {
			execution, err := dbClient.GetJobExecution(context.Background(), jobExecution.ID)
			assert.NilError(t, err)
			switch jobIDNameMap[execution.JobID] {
			case "job1":
				assert.Equal(t, api.StatusQueued, execution.Status)
			case "job2":
				assert.Equal(t, api.
					StatusCreated, execution.Status)
			}
		}
		t.Run("dep_is_success", func(t *testing.T) {
			for _, execution := range createdPipeline.CreatedJobExecutions {
				if jobIDNameMap[execution.JobID] == "job1" {
					err := dbClient.UpdateJobExecution(context.Background(), &db.UpdateJobExecutionOption{
						ID:     execution.ID,
						Status: lo.ToPtr(api.StatusSucceeded),
					})
					execution.Status = api.StatusSucceeded
					assert.NilError(t, err)
				}
			}
			err := s.Dispatch(context.Background(), createdPipeline.CreatedJobs, createdPipeline.CreatedJobExecutions)
			assert.NilError(t, err)
			for _, jobExecution := range createdPipeline.CreatedJobExecutions {
				execution, err := dbClient.GetJobExecution(context.Background(), jobExecution.ID)
				assert.NilError(t, err)
				switch jobIDNameMap[execution.JobID] {
				case "job1":
					assert.Equal(t, api.StatusSucceeded, execution.Status)
				case "job2":
					assert.Equal(t, api.StatusQueued, execution.Status)
				}
			}
		})
		t.Run("dep_is_not_success", func(t *testing.T) {
			for _, execution := range createdPipeline.CreatedJobExecutions {
				if jobIDNameMap[execution.JobID] == "job1" {
					err := dbClient.UpdateJobExecution(context.Background(), &db.UpdateJobExecutionOption{
						ID:     execution.ID,
						Status: lo.ToPtr(api.StatusFailed),
					})
					execution.Status = api.StatusFailed
					assert.NilError(t, err)
				}
			}
			err := s.Dispatch(context.Background(), createdPipeline.CreatedJobs, createdPipeline.CreatedJobExecutions)
			assert.NilError(t, err)
			for _, jobExecution := range createdPipeline.CreatedJobExecutions {
				execution, err := dbClient.GetJobExecution(context.Background(), jobExecution.ID)
				assert.NilError(t, err)
				switch jobIDNameMap[execution.JobID] {
				case "job1":
					assert.Equal(t, api.StatusFailed, execution.Status)
				case "job2":
					assert.Equal(t, api.StatusSkipped, execution.Status)
				}
			}
		})
	})
}
