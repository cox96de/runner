package handler

import (
	"context"
	"testing"

	"github.com/cox96de/runner/entity"
	"github.com/cox96de/runner/mock"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"
)

func TestHandler_createPipeline(t *testing.T) {
	dbClient := mock.NewMockDB(t)
	handler := NewHandler(dbClient)
	pipeline, err := handler.createPipeline(context.Background(), &entity.Pipeline{
		Jobs: []*entity.Job{
			{
				Name:             "job1",
				RunsOn:           &entity.RunsOn{Label: "label1"},
				WorkingDirectory: "/tmp",
				EnvVar:           map[string]string{"key": "value"},
				DependsOn:        []string{"job2"},
				Steps: []*entity.Step{
					{
						Name:     "step1",
						User:     "user",
						Commands: []string{"echo hello"},
					},
				},
			},
		},
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, pipeline, &entity.Pipeline{
		Jobs: []*entity.Job{
			{
				Name:             "job1",
				RunsOn:           &entity.RunsOn{Label: "label1"},
				WorkingDirectory: "/tmp",
				EnvVar:           map[string]string{"key": "value"},
				DependsOn:        []string{"job2"},
				Executions: []*entity.JobExecution{
					{
						Status: entity.JobStatusCreated,
					},
				},
				Steps: []*entity.Step{
					{
						Name:     "step1",
						User:     "user",
						Commands: []string{"echo hello"},
						Executions: []*entity.StepExecution{
							{
								Status: entity.StepStatusCreated,
							},
						},
					},
				},
			},
		},
	}, cmpopts.IgnoreFields(entity.Pipeline{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(entity.Job{}, "PipelineID", "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(entity.JobExecution{}, "JobID", "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"),
		cmpopts.IgnoreFields(entity.Step{}, "PipelineID", "JobID", "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(entity.StepExecution{}, "JobExecutionID", "ID", "CreatedAt", "UpdatedAt", "StartedAt", "CompletedAt"))
}
