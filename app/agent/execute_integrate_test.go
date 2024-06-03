package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"gotest.tools/v3/fs"

	"github.com/cox96de/runner/api/httpserverclient"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/handler"
	"github.com/cox96de/runner/app/server/logstorage"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/mock"
	"github.com/gin-gonic/gin"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/engine/shell"
	"gotest.tools/v3/assert"
)

func newMockServerHandler(t *testing.T) *httpserverclient.Client {
	dbClient := mock.NewMockDB(t)
	h := handler.NewHandler(dbClient, pipeline.NewService(dbClient), dispatch.NewService(dbClient), mock.NewMockLocker(),
		logstorage.NewService(mock.NewMockRedis(t), logstorage.NewFilesystemOSS(fs.NewDir(t, "baseDir").Path())))
	engine := gin.New()
	h.RegisterRouter(engine.Group("/api/v1"))
	server := httptest.NewServer(engine)
	client, err := httpserverclient.NewClient(&http.Client{}, server.URL)
	assert.NilError(t, err)
	return client
}

func TestExecution(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip test on windows")
		}
		client := newMockServerHandler(t)
		ctx := context.Background()
		_, err := client.CreatePipeline(ctx, &api.CreatePipelineRequest{
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
		requestJobResponse, err := client.RequestJob(ctx, &api.RequestJobRequest{})
		assert.NilError(t, err)
		execution := NewExecution(shell.NewEngine(), requestJobResponse.Job, client)
		assert.Assert(t, execution != nil)
		err = execution.Execute(ctx)
		assert.NilError(t, err)
	})
	t.Run("timeout", func(t *testing.T) {
		t.Skip("FIXME: timeout status")
		if runtime.GOOS == "windows" {
			t.Skip("skip test on windows")
		}
		client := newMockServerHandler(t)
		ctx := context.Background()
		_, err := client.CreatePipeline(ctx, &api.CreatePipelineRequest{
			Pipeline: &api.PipelineDSL{
				Jobs: []*api.JobDSL{{
					Name:    "job1",
					Timeout: 1,
					Steps: []*api.StepDSL{{
						Name:     "step1",
						Commands: []string{"sleep 10"},
					}},
				}},
			},
		})
		assert.NilError(t, err)
		requestJobResponse, err := client.RequestJob(ctx, &api.RequestJobRequest{})
		assert.NilError(t, err)
		execution := NewExecution(shell.NewEngine(), requestJobResponse.Job, client)
		assert.Assert(t, execution != nil)
		err = execution.Execute(ctx)
		assert.NilError(t, err)
		executions, err := client.ListJobExecutions(ctx, &api.ListJobExecutionsRequest{
			JobID: requestJobResponse.Job.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, executions.Jobs[0].Status, api.StatusSucceeded)
	})
	t.Run("step_failed", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip test on windows")
		}
		client := newMockServerHandler(t)
		ctx := context.Background()
		_, err := client.CreatePipeline(ctx, &api.CreatePipelineRequest{
			Pipeline: &api.PipelineDSL{
				Jobs: []*api.JobDSL{{
					Name:    "job1",
					Timeout: 1,
					Steps: []*api.StepDSL{
						{
							Name:     "step1",
							Commands: []string{"echo hello"},
						},
						{
							Name:     "step2",
							Commands: []string{"exit 1"},
						},
					},
				}},
			},
		})
		assert.NilError(t, err)
		requestJobResponse, err := client.RequestJob(ctx, &api.RequestJobRequest{})
		assert.NilError(t, err)
		execution := NewExecution(shell.NewEngine(), requestJobResponse.Job, client)
		assert.Assert(t, execution != nil)
		err = execution.Execute(ctx)
		assert.NilError(t, err)
		executions, err := client.ListJobExecutions(ctx, &api.ListJobExecutionsRequest{
			JobID: requestJobResponse.Job.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, executions.Jobs[0].Status, api.StatusFailed)
	})
	t.Run("bad_dag", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip test on windows")
		}
		client := newMockServerHandler(t)
		ctx := context.Background()
		_, err := client.CreatePipeline(ctx, &api.CreatePipelineRequest{
			Pipeline: &api.PipelineDSL{
				Jobs: []*api.JobDSL{{
					Name:    "job1",
					Timeout: 1,
					Steps: []*api.StepDSL{
						{
							Name:      "step1",
							Commands:  []string{"echo step1"},
							DependsOn: []string{"step3"},
						},
						{
							Name:      "step2",
							Commands:  []string{"echo step2"},
							DependsOn: []string{"step1"},
						},
						{
							Name:      "step3",
							Commands:  []string{"echo step3"},
							DependsOn: []string{"step2"},
						},
					},
				}},
			},
		})
		assert.NilError(t, err)
		requestJobResponse, err := client.RequestJob(ctx, &api.RequestJobRequest{})
		assert.NilError(t, err)
		execution := NewExecution(shell.NewEngine(), requestJobResponse.Job, client)
		assert.Assert(t, execution != nil)
		err = execution.Execute(ctx)
		assert.NilError(t, err)
		executions, err := client.ListJobExecutions(ctx, &api.ListJobExecutionsRequest{
			JobID: requestJobResponse.Job.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, executions.Jobs[0].Status, api.StatusFailed)
	})
}
