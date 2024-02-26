package httpserverclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samber/lo"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/handler"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/mock"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func TestNewClient(t *testing.T) {
	dbClient := mock.NewMockDB(t)
	pipelineService := pipeline.NewService(dbClient)
	dispatchService := dispatch.NewService(dbClient)
	locker := mock.NewMockLocker()
	h := handler.NewHandler(dbClient, pipelineService, dispatchService, locker)
	engine := gin.New()
	h.RegisterRouter(engine.Group("/api/v1"))
	server := httptest.NewServer(engine)
	client, err := NewClient(&http.Client{}, server.URL)
	assert.NilError(t, err)
	ctx := context.Background()
	createPipelineResponse, err := client.CreatePipeline(ctx, &api.CreatePipelineRequest{
		Pipeline: &api.PipelineDSL{
			Jobs: []*api.JobDSL{{
				Name: "job1",
				Steps: []*api.StepDSL{{
					Name:     "step1",
					Commands: []string{"echo hello"},
				}},
			}},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, len(createPipelineResponse.Pipeline.Jobs) == 1)
	t.Run("RequestJob", func(t *testing.T) {
		requestJobResponse, err := client.RequestJob(ctx, &api.RequestJobRequest{})
		assert.NilError(t, err)
		assert.Assert(t, requestJobResponse.Job != nil)
		requestedJob := requestJobResponse.Job
		requestJobResponse, err = client.RequestJob(ctx, &api.RequestJobRequest{})
		assert.NilError(t, err)
		assert.Assert(t, requestJobResponse.Job == nil)
		t.Run("UpdateJobExecution", func(t *testing.T) {
			updateJobExecutionResponse, err := client.UpdateJobExecution(ctx, &api.UpdateJobExecutionRequest{
				JobID:          requestedJob.ID,
				JobExecutionID: requestedJob.Executions[0].ID,
				Status:         lo.ToPtr(api.StatusPreparing),
			})
			assert.NilError(t, err)
			assert.Assert(t, updateJobExecutionResponse.Job.Status == api.StatusPreparing)
		})
	})
}
