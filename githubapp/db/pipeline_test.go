package db

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestClient_CreatePipeline(t *testing.T) {
	db := NewMockDB(t)
	pipeline, err := db.CreatePipeline(context.Background(), &CreatePipelineOption{
		AppInstallID: 1,
		RepoOwner:    "cox96de",
		RepoName:     "runner",
		HeadSHA:      "4c4236e5850e58ee32f71e84e817f48296e56de8",
	})
	assert.NilError(t, err)
	assert.Equal(t, pipeline.AppInstallID, int64(1))
	assert.Equal(t, pipeline.RepoOwner, "cox96de")
	assert.Equal(t, pipeline.RepoName, "runner")
	assert.Equal(t, pipeline.HeadSHA, "4c4236e5850e58ee32f71e84e817f48296e56de8")
	t.Run("get", func(t *testing.T) {
		getPipelineByID, err := db.GetPipelineByID(context.Background(), pipeline.ID)
		assert.NilError(t, err)
		assert.DeepEqual(t, getPipelineByID, pipeline)
	})
}
