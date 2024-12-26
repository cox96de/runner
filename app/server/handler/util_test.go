package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cox96de/runner/api"

	"github.com/cockroachdb/errors"

	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func TestBind(t *testing.T) {
	engine := gin.New()
	type Test struct {
		Name   string `json:"name"`
		Header string `header:"header"`
		Query  string `query:"query"`
		Path   string `path:"path"`
	}
	req := &Test{}
	engine.Any("/test/:path", func(c *gin.Context) {
		if err := Bind(c, req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})
	server := httptest.NewServer(engine)
	defer server.Close()
	request, err := http.NewRequest(http.MethodPost, server.URL+"/test/path?query=query_value", bytes.NewReader([]byte("{\"name\":\"test\"}")))
	assert.NilError(t, err)
	request.Header.Add("header", "header_value")
	request.Header.Add("Content-Type", "application/json")
	assert.NilError(t, err)
	do, err := server.Client().Do(request)
	assert.NilError(t, err)
	assert.Equal(t, do.StatusCode, http.StatusOK)
	assert.DeepEqual(t, req, &Test{
		Name:   "test",
		Header: "header_value",
		Query:  "query_value",
		Path:   "path",
	})
}

func Test_getGinHandler(t *testing.T) {
	engine := gin.New()
	engine.Any("/ping", getGinHandler(func(ctx context.Context, request *string) (*string, error) {
		return nil, &HTTPError{Code: 444, CauseError: errors.New(t.Name())}
	}))
	server := httptest.NewServer(engine)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/ping", nil)
	assert.NilError(t, err)
	do, err := server.Client().Do(request)
	assert.NilError(t, err)
	assert.Equal(t, do.StatusCode, 444)
}

func CreateAndPushToStatus(t *testing.T, h *Handler, pipeline *api.PipelineDSL, targetStatus api.Status) *api.Job {
	ctx := context.Background()
	createPipelineResponse, err := h.CreatePipeline(ctx, &api.CreatePipelineRequest{Pipeline: pipeline})
	assert.NilError(t, err)
	jobExecutionID := createPipelineResponse.Pipeline.Jobs[0].Execution.ID
	PushJobToStatus(t, h, ctx, jobExecutionID, api.StatusCreated, targetStatus)
	return createPipelineResponse.Pipeline.Jobs[0]
}

func PushJobToStatus(t *testing.T, h *Handler, ctx context.Context, jobExecutionID int64, currentStatus, targetStatus api.Status) api.Status {
	if currentStatus >= targetStatus {
		return currentStatus
	}
	switch {
	case targetStatus == api.StatusCreated:
	case targetStatus == api.StatusQueued:
	case targetStatus == api.StatusPreparing:
		currentStatus = PushJobToStatus(t, h, ctx, jobExecutionID, currentStatus, api.StatusQueued)
	case targetStatus == api.StatusRunning:
		currentStatus = PushJobToStatus(t, h, ctx, jobExecutionID, currentStatus, api.StatusPreparing)
	case targetStatus.IsCompleted():
		currentStatus = PushJobToStatus(t, h, ctx, jobExecutionID, currentStatus, api.StatusRunning)
	case targetStatus == api.StatusCanceling:
	}
	_, err := h.UpdateJobExecution(ctx, &api.UpdateJobExecutionRequest{
		JobExecutionID: jobExecutionID,
		Status:         &targetStatus,
	})
	assert.NilError(t, err)
	return currentStatus
}
