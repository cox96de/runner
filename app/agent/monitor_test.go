package agent

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	mockapi "github.com/cox96de/runner/api/mock"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"gotest.tools/v3/assert"
)

func TestExecution_monitorHeartbeat(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		client := mockapi.NewMockServerClient(gomock.NewController(t))
		jobExecutionID := int64(1)
		client.EXPECT().Heartbeat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
			request *api.HeartbeatRequest, option ...grpc.CallOption,
		) (*api.HeartbeatResponse, error) {
			assert.Equal(t, jobExecutionID, request.JobExecutionID)
			return &api.HeartbeatResponse{}, nil
		}).AnyTimes()
		e := &Execution{
			engine:           nil,
			job:              &api.Job{},
			jobExecution:     &api.JobExecution{ID: jobExecutionID},
			stepExecutions:   nil,
			client:           client,
			logFlushInternal: 0,
			runner:           nil,
			dag:              nil,
			logWriter:        nil,
			jobCtx:           nil,
			jobCanceller:     nil,
			abortedReason:    atomic.Uint32{},
		}
		e.jobCtx, e.jobCanceller = context.WithCancel(context.Background())
		go e.monitorHeartbeat(context.Background(), time.Millisecond, time.Millisecond*4)
		time.Sleep(time.Millisecond * 20)
		assert.Assert(t, e.jobCtx.Err() == nil)
	})
	t.Run("timeout", func(t *testing.T) {
		client := mockapi.NewMockServerClient(gomock.NewController(t))
		jobExecutionID := int64(1)
		client.EXPECT().Heartbeat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
			request *api.HeartbeatRequest, option ...grpc.CallOption,
		) (*api.HeartbeatResponse, error) {
			assert.Equal(t, jobExecutionID, request.JobExecutionID)
			return nil, errors.New("wrong")
		}).AnyTimes()
		e := &Execution{
			engine:           nil,
			job:              &api.Job{},
			jobExecution:     &api.JobExecution{ID: jobExecutionID},
			stepExecutions:   nil,
			client:           client,
			logFlushInternal: 0,
			runner:           nil,
			dag:              nil,
			logWriter:        nil,
			jobCtx:           nil,
			jobCanceller:     nil,
			abortedReason:    atomic.Uint32{},
		}
		e.jobCtx, e.jobCanceller = context.WithCancel(context.Background())
		go e.monitorHeartbeat(context.Background(), time.Millisecond, time.Millisecond*4)
		time.Sleep(time.Millisecond * 10)
		assert.Assert(t, e.jobCtx.Err() != nil)
	})
}
