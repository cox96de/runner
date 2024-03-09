package shell

import (
	"context"
	"runtime"
	"testing"

	"github.com/cox96de/runner/app/executor/executorpb"

	"gotest.tools/v3/assert"
)

func TestRunner_Start(t *testing.T) {
	r := NewRunner()
	ctx := context.Background()
	err := r.Start(ctx)
	assert.NilError(t, err)
	e, err := r.GetExecutor(ctx)
	assert.NilError(t, err)
	_, err = e.Ping(ctx, &executorpb.PingRequest{})
	assert.NilError(t, err)
	var startCommandResp *executorpb.StartCommandResponse
	if runtime.GOOS == "windows" {
		startCommandResp, err = e.StartCommand(ctx, &executorpb.StartCommandRequest{
			Dir:      "C:\\",
			Commands: []string{"cmd", "/c", "dir"},
		})
	} else {
		startCommandResp, err = e.StartCommand(ctx, &executorpb.StartCommandRequest{
			Dir:      "/tmp",
			Commands: []string{"ls"},
		})
	}
	assert.NilError(t, err)
	getCommandLogResp, err := e.GetCommandLog(ctx, &executorpb.GetCommandLogRequest{Pid: startCommandResp.Status.Pid})
	assert.NilError(t, err)
	all, err := executorpb.ReadAllFromCommandLog(getCommandLogResp)
	assert.NilError(t, err)
	t.Logf("%s", all)
	t.Run("stop", func(t *testing.T) {
		err := r.Stop(ctx)
		assert.NilError(t, err)
	})
}
