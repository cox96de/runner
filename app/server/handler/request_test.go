package handler

import (
	"context"
	"testing"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/mock"
	"gotest.tools/v3/assert"
)

func TestHandler_requestJobHandler(t *testing.T) {
	dbClient := mock.NewMockDB(t)
	handler := NewHandler(dbClient, pipeline.NewService(dbClient), dispatch.NewService(dbClient), mock.NewMockLocker())
	createPipelineResponse, err := handler.CreatePipeline(context.Background(), &api.CreatePipelineRequest{
		Pipeline: &api.PipelineDSL{
			Jobs: []*api.JobDSL{
				{
					Name: "test",
					Steps: []*api.StepDSL{
						{
							Name:     "test",
							Commands: []string{"echo", "test"},
						},
					},
				},
			},
		},
	})
	assert.NilError(t, err)
	response, err := handler.RequestJob(context.Background(), &api.RequestJobRequest{})
	assert.NilError(t, err)
	job := response.Job
	assert.Equal(t, job.ID, createPipelineResponse.Pipeline.Jobs[0].ID)
	assert.Equal(t, job.Executions[0].ID, createPipelineResponse.Pipeline.Jobs[0].Executions[0].ID)
	t.Run("get_empty_job", func(t *testing.T) {
		requestJobResponse, err := handler.RequestJob(context.Background(), &api.RequestJobRequest{})
		assert.NilError(t, err)
		assert.Assert(t, requestJobResponse.Job == nil)
	})
}
