package handler

import (
	"context"
	"testing"

	"github.com/cox96de/runner/app/server/logstorage"
	pipeline2 "github.com/cox96de/runner/app/server/pipeline"
	"gotest.tools/v3/fs"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/eventhook"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/mock"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

func TestHandler_UpdateJobExecution(t *testing.T) {
	dbCli := mock.NewMockDB(t)
	eventHook := eventhook.NewService(eventhook.NewNopSender())
	handler := NewHandler(dbCli, nil, dispatch.NewService(dbCli, eventHook), mock.NewMockLocker(),
		nil, eventHook)
	jobs, err := handler.db.CreateJobs(context.Background(), []*db.CreateJobOption{
		{
			PipelineID: 1,
		},
	})
	assert.NilError(t, err)
	job := jobs[0]
	executions, err := handler.db.CreateJobExecutions(context.Background(), []*db.CreateJobExecutionOption{
		{
			JobID:  job.ID,
			Status: api.StatusCreated,
		},
	})
	assert.NilError(t, err)
	t.Run("bad_status_transmit", func(t *testing.T) {
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobExecutionID: executions[0].ID,
			Status:         lo.ToPtr(api.StatusPreparing),
		})
		assert.Assert(t, err != nil)
	})
	t.Run("transmit", func(t *testing.T) {
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobExecutionID: executions[0].ID,
			Status:         lo.ToPtr(api.StatusQueued),
		})
		assert.NilError(t, err)
		execution, err := handler.db.GetJobExecution(context.Background(), executions[0].ID)
		assert.NilError(t, err)
		assert.DeepEqual(t, execution.Status, api.StatusQueued)
	})
	t.Run("ListJobExecutions", func(t *testing.T) {
		listJobExecutionsResponse, err := handler.ListJobExecutions(context.Background(), &api.ListJobExecutionsRequest{
			JobID: 1,
		})
		assert.NilError(t, err)
		assert.Assert(t, len(listJobExecutionsResponse.Jobs) == 1)
		assert.Assert(t, listJobExecutionsResponse.Jobs[0].ID == executions[0].ID)
	})
	t.Run("GetJobExecution", func(t *testing.T) {
		jobExecutionResponse, err := handler.GetJobExecution(context.Background(), &api.GetJobExecutionRequest{
			JobExecutionID: executions[0].ID,
		})
		assert.NilError(t, err)
		assert.Assert(t, jobExecutionResponse.JobExecution.ID == executions[0].ID)
	})
	t.Run("ToCancelling", func(t *testing.T) {
		executions, err := handler.db.CreateJobExecutions(context.Background(), []*db.CreateJobExecutionOption{
			{
				JobID:  job.ID,
				Status: api.StatusCreated,
			},
		})
		assert.NilError(t, err)
		jobExecution := executions[0]
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobExecutionID: jobExecution.ID,
			Status:         lo.ToPtr(api.StatusQueued),
			Reason:         nil,
		})
		assert.NilError(t, err)
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobExecutionID: jobExecution.ID,
			Status:         lo.ToPtr(api.StatusPreparing),
			Reason:         nil,
		})
		assert.NilError(t, err)
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobExecutionID: jobExecution.ID,
			Status:         lo.ToPtr(api.StatusRunning),
			Reason:         nil,
		})
		assert.NilError(t, err)
		_, err = handler.UpdateJobExecution(context.Background(), &api.UpdateJobExecutionRequest{
			JobExecutionID: jobExecution.ID,
			Status:         lo.ToPtr(api.StatusCanceling),
			Reason:         nil,
		})
		assert.NilError(t, err)
	})
}

func TestHandler_RerunJob(t *testing.T) {
	t.Run("rerun", func(t *testing.T) {
		dbCli := mock.NewMockDB(t)
		eventHook := eventhook.NewService(eventhook.NewNopSender())
		mockLogService := logstorage.NewService(mock.NewMockRedis(t), logstorage.NewFilesystemOSS(fs.NewDir(t, "baseDir").Path()))
		handler := NewHandler(dbCli, pipeline2.NewService(dbCli), dispatch.NewService(dbCli, eventHook),
			mock.NewMockLocker(), mockLogService, eventHook)
		createPipelineResp, err := handler.CreatePipeline(context.Background(), &api.CreatePipelineRequest{Pipeline: &api.PipelineDSL{Jobs: []*api.JobDSL{
			{
				Name: "Job",
				RunsOn: &api.RunsOn{
					Label: "label",
				},
				WorkingDirectory: "",
				EnvVar:           nil,
				DependsOn:        nil,
				Steps:            []*api.StepDSL{{Name: "step", Commands: []string{"command"}}},
				Timeout:          0,
			},
		}}})
		assert.NilError(t, err)
		job := createPipelineResp.Pipeline.Jobs[0]

		jobExecutions, err := dbCli.GetJobExecutionsByJobID(context.Background(), job.ID)
		jobExecution := jobExecutions[len(jobExecutions)-1]
		PushJobToStatus(t, handler, context.Background(), jobExecution.ID, jobExecution.Status, api.StatusFailed)
		assert.NilError(t, err)
		rerunJobResponse, err := handler.RerunJob(context.Background(), &api.RerunJobRequest{
			JobID: job.ID,
		})
		assert.NilError(t, err)
		assert.Assert(t, rerunJobResponse.JobExecution.JobID == job.ID)
		assert.DeepEqual(t, rerunJobResponse.JobExecution.Status, api.StatusCreated)
		for _, step := range rerunJobResponse.JobExecution.Steps {
			assert.DeepEqual(t, step.Status, api.StatusCreated)
		}
	})
}
