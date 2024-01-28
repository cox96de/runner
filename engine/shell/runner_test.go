package shell

import (
	"context"
	"io"
	"runtime"
	"testing"

	"github.com/cox96de/runner/internal/executor"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Start(t *testing.T) {
	if runtime.GOOS == "windows" {
		// TODO: support on windows.
		t.Skipf("windows test is not supported yet")
	}
	r := NewRunner()
	ctx := context.Background()
	err := r.Start(ctx)
	assert.NoError(t, err)
	e, err := r.GetExecutor(ctx, "")
	assert.NoError(t, err)
	err = e.Ping(ctx)
	assert.NoError(t, err)
	err = e.StartCommand(ctx, "a", &executor.StartCommandRequest{
		Dir:     "/tmp",
		Command: []string{"ls"},
	})
	assert.NoError(t, err)
	logReader := e.GetCommandLogs(ctx, "a")
	all, err := io.ReadAll(logReader)
	assert.NoError(t, err)
	t.Logf("%s", string(all))
	t.Run("stop", func(t *testing.T) {
		err := r.Stop(ctx)
		assert.NoError(t, err)
	})
}
