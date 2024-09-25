package monitor

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cox96de/runner/app/server/logstorage"
	"gotest.tools/v3/fs"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/mock"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

func TestService_RecycleHeartbeatTimeoutJobs(t *testing.T) {
	dbCli := mock.NewMockDB(t)
	pipelineService := pipeline.NewService(dbCli)
	createPipelineResponse, err := pipelineService.CreatePipeline(context.Background(), &api.PipelineDSL{Jobs: []*api.JobDSL{
		{
			Steps: []*api.StepDSL{
				{},
				{},
			},
		},
	}})
	assert.NilError(t, err)
	jobExecution := createPipelineResponse.CreatedJobExecutions[0]
	err = dispatch.UpdateJobExecution(context.Background(), dbCli, &db.UpdateJobExecutionOption{
		ID:          jobExecution.ID,
		Status:      lo.ToPtr(api.StatusQueued),
		Reason:      nil,
		StartedAt:   nil,
		CompletedAt: nil,
	})
	assert.NilError(t, err)
	err = dispatch.UpdateJobExecution(context.Background(), dbCli, &db.UpdateJobExecutionOption{
		ID:          jobExecution.ID,
		Status:      lo.ToPtr(api.StatusPreparing),
		Reason:      nil,
		StartedAt:   nil,
		CompletedAt: nil,
	})
	assert.NilError(t, err)
	err = dbCli.UpdateStepExecution(context.Background(),
		&db.UpdateStepExecutionOption{
			ID:     createPipelineResponse.CreatedStepExecutions[0].ID,
			Status: lo.ToPtr(api.StatusSucceeded),
		})
	assert.NilError(t, err)
	err = dbCli.TouchHeartbeat(context.Background(), jobExecution.ID)
	assert.NilError(t, err)
	s := NewService(dbCli, logstorage.NewService(mock.NewMockRedis(t), logstorage.NewFilesystemOSS(fs.NewDir(t, "baseDir").Path())))
	t.Run("not_timeout", func(t *testing.T) {
		err = s.RecycleHeartbeatTimeoutJobs(context.Background(), time.Second)
		assert.NilError(t, err)
		execution, err := dbCli.GetJobExecution(context.Background(), jobExecution.ID)
		assert.NilError(t, err)
		assert.Assert(t, execution.Status.IsRunning())
	})
	time.Sleep(time.Millisecond * 5)
	err = s.RecycleHeartbeatTimeoutJobs(context.Background(), time.Millisecond)
	assert.NilError(t, err)
	execution, err := dbCli.GetJobExecution(context.Background(), jobExecution.ID)
	assert.NilError(t, err)
	assert.Equal(t, execution.Status, api.StatusFailed)
	reason := api.Reason{}
	_ = json.Unmarshal(execution.Reason, &reason)
	assert.Equal(t, reason.Reason, api.FailedReasonHeartbeatTimeout)
	stepExecutions, err := dbCli.GetStepExecutionsByJobExecutionID(context.Background(), jobExecution.ID)
	assert.NilError(t, err)
	for _, stepExecution := range stepExecutions {
		assert.Assert(t, stepExecution.Status == api.StatusSkipped || stepExecution.Status == api.StatusSucceeded)
	}
}
