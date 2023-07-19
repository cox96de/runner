package engine

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
)

func TestWaitExecutorReady(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		executor := NewMockExecutor(gomock.NewController(t))
		executor.EXPECT().Ping(gomock.Any()).Return(nil)
		err := WaitExecutorReady(context.Background(), executor, time.Second, time.Hour)
		assert.NilError(t, err)
	})
	t.Run("context done", func(t *testing.T) {
		executor := NewMockExecutor(gomock.NewController(t))
		executor.EXPECT().Ping(gomock.Any()).Return(errors.New("some error")).AnyTimes()
		err := WaitExecutorReady(context.Background(), executor, time.Millisecond*10, time.Millisecond*20)
		assert.ErrorContains(t, err, "context deadline exceeded")
	})
	t.Run("second ping success", func(t *testing.T) {
		executor := NewMockExecutor(gomock.NewController(t))
		executor.EXPECT().Ping(gomock.Any()).Return(errors.New("some error")).Times(1)
		executor.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)
		err := WaitExecutorReady(context.Background(), executor, time.Millisecond*5, time.Hour)
		assert.NilError(t, err)
	})
	t.Run("0 interval", func(t *testing.T) {
		executor := NewMockExecutor(gomock.NewController(t))
		err := WaitExecutorReady(context.Background(), executor, 0, time.Hour)
		assert.ErrorContains(t, err, "interval must be greater than 0")
	})
}
