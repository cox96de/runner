package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cox96de/runner/api"
	mockapi "github.com/cox96de/runner/api/mock"
	"github.com/cox96de/runner/log"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"gotest.tools/v3/assert"
)

func Test_newLogCollector(t *testing.T) {
	getMockServerClient := func(t *testing.T, logs *[]*api.LogLine, lock *sync.Mutex) *mockapi.MockServerClient {
		mockServerClient := mockapi.NewMockServerClient(gomock.NewController(t))
		mockServerClient.EXPECT().UploadLogLines(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, req *api.UpdateLogLinesRequest, _ ...grpc.CallOption) (*api.UpdateLogLinesResponse, error) {
				lock.Lock()
				defer lock.Unlock()
				*logs = append(*logs, req.Lines...)
				return nil, nil
			}).AnyTimes()
		return mockServerClient
	}
	t.Run("simple", func(t *testing.T) {
		logName := t.Name()
		var (
			logs []*api.LogLine
			l    sync.Mutex
		)
		mockServerClient := getMockServerClient(t, &logs, &l)
		flushInterval := time.Millisecond * 10
		collector := newLogCollector(mockServerClient, &api.JobExecution{}, logName, log.ExtractLogger(context.Background()), flushInterval)
		_, err := collector.Write([]byte("a\nb\n"))
		assert.NilError(t, err)
		time.Sleep(flushInterval * 2)
		err = collector.Close()
		assert.NilError(t, err)
		validateLogs(t, logs, []string{"a", "b"})
	})
	t.Run("normal", func(t *testing.T) {
		logName := t.Name()
		var (
			logs []*api.LogLine
			l    sync.Mutex
		)
		mockServerClient := getMockServerClient(t, &logs, &l)
		flushInterval := time.Millisecond * 10
		collector := newLogCollector(mockServerClient, &api.JobExecution{}, logName, log.ExtractLogger(context.Background()), flushInterval)
		_, err := collector.Write([]byte("a\nb\nc"))
		assert.NilError(t, err)
		time.Sleep(flushInterval * 3)
		l.Lock()
		validateLogs(t, logs, []string{"a", "b"})
		l.Unlock()
		_, err = collector.Write([]byte("d\ne"))
		assert.NilError(t, err)
		time.Sleep(flushInterval * 3)
		l.Lock()
		validateLogs(t, logs, []string{"a", "b", "cd"})
		l.Unlock()

		err = collector.Close()
		assert.NilError(t, err)
		l.Lock()
		validateLogs(t, logs, []string{"a", "b", "cd", "e"})
		l.Unlock()
	})
	t.Run("windows", func(t *testing.T) {
		logName := t.Name()
		var (
			logs []*api.LogLine
			l    sync.Mutex
		)
		mockServerClient := getMockServerClient(t, &logs, &l)
		flushInterval := time.Millisecond * 10
		collector := newLogCollector(mockServerClient, &api.JobExecution{}, logName, log.ExtractLogger(context.Background()), flushInterval)
		_, err := collector.Write([]byte("\r\na\nb\rc"))
		assert.NilError(t, err)
		time.Sleep(flushInterval * 2)
		l.Lock()
		validateLogs(t, logs, []string{"", "a", "b"})
		l.Unlock()
		_, err = collector.Write([]byte("d\re"))
		assert.NilError(t, err)
		time.Sleep(flushInterval * 2)
		l.Lock()
		validateLogs(t, logs, []string{"", "a", "b", "cd"})
		l.Unlock()

		err = collector.Close()
		assert.NilError(t, err)
		l.Lock()
		validateLogs(t, logs, []string{"", "a", "b", "cd", "e"})
		l.Unlock()
	})
}

func validateLogs(t *testing.T, loglines []*api.LogLine, expected []string) {
	t.Helper()
	assert.Equal(t, len(loglines), len(expected))
	if len(loglines) == 0 {
		return
	}
	time := loglines[0].Timestamp
	offset := loglines[0].Number
	for i, line := range loglines {
		assert.Assert(t, line.Timestamp >= time, i)
		assert.Equal(t, line.Number, int64(i)+offset, i)
		assert.Equal(t, line.Output, expected[i])
	}
}

func Test_logCollector_Close(t *testing.T) {
	t.Run("multiple_close", func(t *testing.T) {
		collector := newLogCollector(nil, &api.JobExecution{}, "", nil, 0)
		err := collector.Close()
		assert.NilError(t, err)
		err = collector.Close()
		assert.ErrorContains(t, err, "log already closed")
	})
	t.Run("retry_to_flush", func(t *testing.T) {
		mockServerClient := mockapi.NewMockServerClient(gomock.NewController(t))
		mockServerClient.EXPECT().UploadLogLines(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("something error")).AnyTimes()
		collector := newLogCollector(mockServerClient, &api.JobExecution{}, t.Name(), log.ExtractLogger(context.Background()), 0)
		_, err := collector.Write([]byte("abcd"))
		assert.NilError(t, err)
		err = collector.Close()
		assert.ErrorContains(t, err, "something error")
	})
}
