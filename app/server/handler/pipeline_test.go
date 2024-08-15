package handler

import (
	"context"
	"testing"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"

	"github.com/cox96de/runner/mock"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"
)

func TestHandler_createPipeline(t *testing.T) {
	dbClient := mock.NewMockDB(t)
	handler := NewHandler(dbClient, pipeline.NewService(dbClient), dispatch.NewService(dbClient), nil, nil)
	p, err := handler.CreatePipeline(context.Background(), &api.CreatePipelineRequest{
		Pipeline: &api.PipelineDSL{
			Jobs: []*api.JobDSL{
				{
					Name:             "job1",
					RunsOn:           &api.RunsOn{Label: "label1"},
					WorkingDirectory: "/tmp",
					EnvVar:           map[string]string{"key": "value"},
					DependsOn:        []string{"job2"},
					Steps: []*api.StepDSL{
						{
							Name:     "step1",
							User:     "user",
							Commands: []string{"echo hello"},
						},
					},
				},
			},
		},
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, p.Pipeline, &api.Pipeline{
		Jobs: []*api.Job{
			{
				Name:             "job1",
				RunsOn:           &api.RunsOn{Label: "label1"},
				WorkingDirectory: "/tmp",
				EnvVar:           map[string]string{"key": "value"},
				DependsOn:        []string{"job2"},
				Execution: &api.JobExecution{
					Status: api.StatusCreated,
					Steps: []*api.StepExecution{
						{
							Status: api.StatusCreated,
						},
					},
				},
				Steps: []*api.Step{
					{
						Name:     "step1",
						User:     "user",
						Commands: []string{"echo hello"},
						Executions: []*api.StepExecution{
							{
								Status: api.StatusCreated,
							},
						},
						Execution: &api.StepExecution{
							Status: api.StatusCreated,
						},
					},
				},
			},
		},
	}, cmpopts.IgnoreUnexported(api.Pipeline{}, api.Job{}, api.JobExecution{}, api.Step{}, api.StepExecution{}, api.RunsOn{}),
		cmpopts.IgnoreFields(api.Pipeline{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(api.Job{}, "PipelineID", "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(api.JobExecution{}, "JobID", "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"),
		cmpopts.IgnoreFields(api.Step{}, "PipelineID", "JobID", "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(api.StepExecution{}, "JobExecutionID", "StepID", "ID", "CreatedAt", "UpdatedAt",
			"StartedAt", "CompletedAt"))
	// Check packing pipeline.
	for _, job := range p.Pipeline.Jobs {
		assert.Equal(t, job.PipelineID, p.Pipeline.ID)
		for _, step := range job.Steps {
			assert.Equal(t, step.PipelineID, p.Pipeline.ID)
			assert.Equal(t, step.JobID, job.ID)
			for _, stepExecution := range step.Executions {
				assert.Equal(t, stepExecution.StepID, step.ID)
			}
		}
		for _, jobExecution := range job.Executions {
			assert.Equal(t, jobExecution.JobID, job.ID)
			for _, stepExecution := range jobExecution.Steps {
				assert.Equal(t, stepExecution.JobExecutionID, jobExecution.ID)
			}
		}
	}
}
