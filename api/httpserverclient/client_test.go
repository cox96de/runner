package httpserverclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/cox96de/runner/app/server/logstorage"
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
	redis := mock.NewMockRedis(t)
	h := handler.NewHandler(dbClient, pipelineService, dispatchService, locker, logstorage.NewService(redis))
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
			assert.Assert(t, updateJobExecutionResponse.JobExecution.Status == api.StatusPreparing)
		})
		t.Run("UploadLogLines", func(t *testing.T) {
			logLines := []*api.LogLine{{
				Timestamp: 0,
				Number:    1,
				Output:    "hello1",
			}}
			updateLogLinesResponse, err := client.UploadLogLines(ctx, &api.UpdateLogLinesRequest{
				JobID:          1,
				JobExecutionID: 1,
				Name:           "step1",
				Lines:          logLines,
			})
			assert.NilError(t, err)
			assert.Assert(t, updateLogLinesResponse != nil)
			t.Run("GetLogLines", func(t *testing.T) {
				getLogLinesResponse, err := client.GetLogLines(ctx, &api.GetLogLinesRequest{
					JobID:          1,
					JobExecutionID: 1,
					Name:           "step1",
					Offset:         0,
					Limit:          lo.ToPtr(int64(1000)),
				})
				assert.NilError(t, err)
				assert.Assert(t, getLogLinesResponse != nil)
				assert.DeepEqual(t, getLogLinesResponse.Lines, logLines, cmpopts.IgnoreUnexported(api.LogLine{}))
			})
		})
		t.Run("UpdateStepExecution", func(t *testing.T) {
			execution, err := client.UpdateStepExecution(ctx, &api.UpdateStepExecutionRequest{
				JobID:           requestedJob.ID,
				JobExecutionID:  requestedJob.Executions[0].ID,
				StepExecutionID: requestedJob.Executions[0].Steps[0].ID,
				Status:          lo.ToPtr(api.StatusRunning),
			})
			assert.NilError(t, err)
			assert.Assert(t, execution != nil)
		})
		t.Run("ListJobExecutions", func(t *testing.T) {
			listJobExecutionsResponse, err := client.ListJobExecutions(ctx, &api.ListJobExecutionsRequest{
				JobID: requestedJob.ID,
			})
			assert.NilError(t, err)
			assert.Assert(t, listJobExecutionsResponse != nil)
			assert.Assert(t, len(listJobExecutionsResponse.Jobs) == 1)
		})
	})
}
