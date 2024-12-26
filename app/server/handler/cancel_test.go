package handler

import (
	"context"
	"testing"

	"github.com/cox96de/runner/app/server/eventhook"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/logstorage"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/mock"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func TestHandler_CancelJobExecution(t *testing.T) {
	ctx := context.Background()
	db := mock.NewMockDB(t)
	eventHook := eventhook.NewService(eventhook.NewNopSender())
	handler := NewHandler(db, pipeline.NewService(db), dispatch.NewService(db, eventHook), mock.NewMockLocker(),
		logstorage.NewService(mock.NewMockRedis(t), logstorage.NewFilesystemOSS(fs.NewDir(t, "test").Path())),
		eventHook)
	t.Run("from_running", func(t *testing.T) {
		job := CreateAndPushToStatus(t, handler, &api.PipelineDSL{
			Jobs: []*api.JobDSL{{
				Name: "job1",
				RunsOn: &api.RunsOn{
					Label: t.Name(),
				},
				Steps: []*api.StepDSL{{
					Name:     "step1",
					Commands: []string{"echo hello"},
				}},
			}},
		}, api.StatusRunning)
		_, err := handler.CancelJobExecution(ctx, &api.CancelJobExecutionRequest{
			JobExecutionID: job.Execution.ID,
		})
		assert.NilError(t, err)
		getJobExecutionResponse, err := handler.GetJobExecution(ctx, &api.GetJobExecutionRequest{JobExecutionID: job.Execution.ID})
		assert.NilError(t, err)
		assert.Equal(t, getJobExecutionResponse.JobExecution.Status, api.StatusCanceling)
	})
	t.Run("pre_running", func(t *testing.T) {
		job := CreateAndPushToStatus(t, handler, &api.PipelineDSL{
			Jobs: []*api.JobDSL{{
				Name: "job1",
				RunsOn: &api.RunsOn{
					Label: t.Name(),
				},
				Steps: []*api.StepDSL{{
					Name:     "step1",
					Commands: []string{"echo hello"},
				}},
			}},
		}, api.StatusQueued)
		_, err := handler.CancelJobExecution(ctx, &api.CancelJobExecutionRequest{
			JobExecutionID: job.Execution.ID,
		})
		assert.NilError(t, err)
		getJobExecutionResponse, err := handler.GetJobExecution(ctx, &api.GetJobExecutionRequest{JobExecutionID: job.Execution.ID})
		assert.NilError(t, err)
		assert.Equal(t, getJobExecutionResponse.JobExecution.Status, api.StatusFailed)
	})
}
