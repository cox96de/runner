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
		job := handler.CreateAndPushToStatus(t, &api.PipelineDSL{
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
		job := handler.CreateAndPushToStatus(t, &api.PipelineDSL{
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

func (h *Handler) CreateAndPushToStatus(t *testing.T, pipeline *api.PipelineDSL, targetStatus api.Status) *api.Job {
	ctx := context.Background()
	createPipelineResponse, err := h.CreatePipeline(ctx, &api.CreatePipelineRequest{Pipeline: pipeline})
	assert.NilError(t, err)
	jobExecutionID := createPipelineResponse.Pipeline.Jobs[0].Execution.ID
	h.PushJobToStatus(t, ctx, jobExecutionID, api.StatusCreated, targetStatus)
	return createPipelineResponse.Pipeline.Jobs[0]
}

func (h *Handler) PushJobToStatus(t *testing.T, ctx context.Context, jobExecutionID int64, currentStatus, targetStatus api.Status) api.Status {
	if currentStatus >= targetStatus {
		return currentStatus
	}
	switch {
	case targetStatus == api.StatusCreated:
	case targetStatus == api.StatusQueued:
	case targetStatus == api.StatusPreparing:
		currentStatus = h.PushJobToStatus(t, ctx, jobExecutionID, currentStatus, api.StatusQueued)
	case targetStatus == api.StatusRunning:
		currentStatus = h.PushJobToStatus(t, ctx, jobExecutionID, currentStatus, api.StatusPreparing)
	case targetStatus.IsCompleted():
		currentStatus = h.PushJobToStatus(t, ctx, jobExecutionID, currentStatus, api.StatusRunning)
	case targetStatus == api.StatusCanceling:
	}
	_, err := h.UpdateJobExecution(ctx, &api.UpdateJobExecutionRequest{
		JobExecutionID: jobExecutionID,
		Status:         &targetStatus,
	})
	assert.NilError(t, err)
	return currentStatus
}
