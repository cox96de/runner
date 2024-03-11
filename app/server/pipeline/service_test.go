package pipeline

import (
	"context"
	"testing"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/mock"
	"gotest.tools/v3/assert"
)

func TestService_CreatePipeline(t *testing.T) {
	service := NewService(mock.NewMockDB(t))
	_, err := service.CreatePipeline(context.Background(), &api.PipelineDSL{
		Jobs: []*api.JobDSL{
			{
				Name: "job1",
				Steps: []*api.StepDSL{
					{
						Name:     "step1",
						Commands: []string{"echo", "hello"},
					},
				},
			},
		},
	})
	assert.NilError(t, err)
}
