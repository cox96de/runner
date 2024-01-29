package shell

import (
	"context"
	"io"
	"runtime"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cox96de/runner/internal/executor"
)

func TestRunner_Start(t *testing.T) {
	r := NewRunner()
	ctx := context.Background()
	err := r.Start(ctx)
	assert.NilError(t, err)
	e, err := r.GetExecutor(ctx, "")
	assert.NilError(t, err)
	err = e.Ping(ctx)
	assert.NilError(t, err)
	if runtime.GOOS == "windows" {
		err = e.StartCommand(ctx, "a", &executor.StartCommandRequest{
			Dir:     "C:\\",
			Command: []string{"cmd", "/c", "dir"},
		})
	} else {
		err = e.StartCommand(ctx, "a", &executor.StartCommandRequest{
			Dir:     "/tmp",
			Command: []string{"ls"},
		})
	}
	assert.NilError(t, err)
	logReader := e.GetCommandLogs(ctx, "a")
	all, err := io.ReadAll(logReader)
	assert.NilError(t, err)
	t.Logf("%s", string(all))
	t.Run("stop", func(t *testing.T) {
		err := r.Stop(ctx)
		assert.NilError(t, err)
	})
}
