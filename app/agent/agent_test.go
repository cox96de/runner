package agent

import (
	"context"
	"testing"

	"github.com/cox96de/runner/api"
	mockapi "github.com/cox96de/runner/api/mock"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
)

func TestAgent_poll(t *testing.T) {
	client := mockapi.NewMockServerClient(gomock.NewController(t))
	dispatchJob := &api.Job{}
	client.EXPECT().RequestJob(gomock.Any(), gomock.Any()).Return(&api.RequestJobResponse{}, nil)
	client.EXPECT().RequestJob(gomock.Any(), gomock.Any()).Return(&api.RequestJobResponse{
		Job: dispatchJob,
	}, nil)
	a := &Agent{
		client: client,
	}
	c := make(chan struct{}, 2)
	c <- struct{}{}
	c <- struct{}{}
	polledJob, err := a.poll(context.Background(), c)
	assert.NilError(t, err)
	assert.Equal(t, dispatchJob, polledJob)
}

func TestAgent_Run(t *testing.T) {
	t.Run("concurrency", func(t *testing.T) {
		// TODO: add test.
	})
}
