package db

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestClient_GetJobByRunnerExecutionID(t *testing.T) {
	db := NewMockDB(t)
	_, err := db.CreateJobs(context.Background(), []*CreateJobOption{
		{
			PipelineID:           1,
			Name:                 "name",
			UID:                  "uid",
			Steps:                nil,
			CheckRunID:           10,
			RunnerJobExecutionID: 11,
		},
	})
	assert.NilError(t, err)
	job, err := db.GetJobByRunnerExecutionID(context.Background(), 11)
	assert.NilError(t, err)
	assert.Equal(t, job.RunnerJobExecutionID, int64(11))
}
