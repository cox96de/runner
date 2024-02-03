package db

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestClient_CreateSteps(t *testing.T) {
	db := NewMockDB(t, &Step{})
	steps, err := db.CreateSteps(context.Background(), []*CreateStepOption{
		{
			PipelineID: 1,
			JobID:      1,
			Name:       "step1",
			User:       "root",
		},
	})
	assert.NilError(t, err)
	for _, step := range steps {
		assert.Assert(t, step.ID > 0, step.Name)
		stepByID, err := db.GetStepByID(context.Background(), step.ID)
		assert.NilError(t, err, step.Name)
		assert.DeepEqual(t, step, stepByID)
	}
}
