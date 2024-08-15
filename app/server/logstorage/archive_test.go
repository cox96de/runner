package logstorage

import (
	"context"
	"strconv"
	"testing"

	"github.com/cox96de/runner/mock"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func TestService_Archive(t *testing.T) {
	dir := fs.NewDir(t, "baseDir")
	service := NewService(mock.NewMockRedis(t), NewFilesystemOSS(dir.Path()))
	err := service.Append(context.Background(), 1, "test", generateTestLog(100))
	assert.NilError(t, err)
	err = service.Archive(context.Background(), 1)
	assert.NilError(t, err)
	t.Run("get_log", func(t *testing.T) {
		logs, err := service.GetLogLines(context.Background(), 1, "test", 0, 100)
		assert.NilError(t, err)
		assert.Equal(t, len(logs), 100)
		for i := 0; i < 100; i++ {
			assert.Equal(t, logs[i].Number, int64(i))
			assert.Equal(t, logs[i].Output, "line "+strconv.Itoa(i))
		}
	})
	t.Run("get_part", func(t *testing.T) {
		logs, err := service.GetLogLines(context.Background(), 1, "test", 1, 90)
		assert.NilError(t, err)
		assert.Equal(t, len(logs), 90)
		for i := 0; i < 90; i++ {
			assert.Equal(t, logs[i].Number, int64(i+1))
			assert.Equal(t, logs[i].Output, "line "+strconv.Itoa(i+1))
		}
	})
}
