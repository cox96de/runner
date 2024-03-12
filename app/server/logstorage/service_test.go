package logstorage

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/mock"
	"gotest.tools/v3/assert"
)

func TestService_Append(t *testing.T) {
	service := NewService(mock.NewMockRedis(t))
	err := service.Append(context.Background(), 1, 1, "test", generateTestLog(100))
	assert.NilError(t, err)
	t.Run("get_less", func(t *testing.T) {
		logs, err := service.GetLogLines(context.Background(), 1, 1, "test", 0, 99)
		assert.NilError(t, err)
		assert.Equal(t, len(logs), 99)
		for i := 0; i < 99; i++ {
			assert.Equal(t, logs[i].Number, int64(i))
			assert.Equal(t, logs[i].Output, fmt.Sprintf("line %d", i))
		}
	})
	t.Run("get_more", func(t *testing.T) {
		logs, err := service.GetLogLines(context.Background(), 1, 1, "test", 0, 1000)
		assert.NilError(t, err)
		assert.Equal(t, len(logs), 100)
		for i := 0; i < 100; i++ {
			assert.Equal(t, logs[i].Number, int64(i))
			assert.Equal(t, logs[i].Output, fmt.Sprintf("line %d", i))
		}
	})
	err = service.Append(context.Background(), 1, 1, "test2", generateTestLog(100))
	assert.NilError(t, err)
	t.Run("log_name_set", func(t *testing.T) {
		logNameSet, err := service.getLogNameSet(context.Background(), 1, 1)
		assert.NilError(t, err)
		assert.DeepEqual(t, logNameSet, []string{"test", "test2"}, cmpopts.SortSlices(func(a, b string) bool {
			return strings.Compare(a, b) < 0
		}))
	})
}

func generateTestLog(lines int) []*api.LogLine {
	logs := make([]*api.LogLine, 0, lines)
	for i := 0; i < lines; i++ {
		logs = append(logs, &api.LogLine{
			Timestamp: int64(i),
			Number:    int64(i),
			Output:    fmt.Sprintf("line %d", i),
		})
	}
	return logs
}
