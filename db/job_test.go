package db

import (
	"context"
	"testing"

	"github.com/cox96de/runner/entity"
	"gotest.tools/v3/assert"
)

func TestClient_CreateJobs(t *testing.T) {
	db := NewMockDB(t, &Job{})
	jobs, err := db.CreateJobs(context.Background(), []*CreateJobOption{
		{
			PipelineID: 1,
			Name:       "job1",
			RunsOn: &entity.RunsOn{
				Label: "label_1",
			},
			WorkingDirectory: "/home/user",
			EnvVar:           map[string]string{"env1": "val1"},
			DependsOn:        []string{"job2"},
		},
		{
			PipelineID: 1,
			Name:       "job2",
			RunsOn: &entity.RunsOn{
				Label: "label_1",
			},
			WorkingDirectory: "/home/user",
			EnvVar:           map[string]string{"env2": "val2"},
		},
	})
	assert.NilError(t, err)
	for _, job := range jobs {
		assert.Assert(t, job.ID > 0, job.Name)
		jobByID, err := db.GetJobByID(context.Background(), job.ID)
		assert.NilError(t, err, job.Name)
		assert.DeepEqual(t, job, jobByID)
	}
}
