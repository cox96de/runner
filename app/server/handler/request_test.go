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
	createPipeline, err := handler.createPipeline(context.Background(), &api.Pipeline{
		Jobs: []*api.Job{
			{
				Name: "test",
				Steps: []*api.Step{
					{
						Name:     "test",
						Commands: []string{"echo", "test"},
					},
				},
			},
		},
	})
	assert.NilError(t, err)
	job, err := handler.requestJobHandler(context.Background())
	assert.NilError(t, err)
	assert.Equal(t, job.ID, createPipeline.Jobs[0].ID)
	assert.Equal(t, job.Executions[0].ID, createPipeline.Jobs[0].Executions[0].ID)
	t.Run("get_empty_job", func(t *testing.T) {
		job, err := handler.requestJobHandler(context.Background())
		assert.NilError(t, err)
		assert.Assert(t, job == nil)
	})
}
