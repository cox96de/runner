package handler

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cox96de/runner/internal/executor"
	"gotest.tools/v3/assert"
)

func TestHandler_startCommandHandler(t *testing.T) {
	testServer, handler := setupHandler(t)
	client := executor.NewClient(testServer.URL)
	dir := os.TempDir()
	t.Run("shot-run", func(t *testing.T) {
		id := t.Name()
		err := client.StartCommand(context.Background(), id,
			&executor.StartCommandRequest{
				Command: []string{"ls", "-alh"},
				Dir:     dir,
			})
		assert.NilError(t, err)
		// TODO: wait for the command to finish
		time.Sleep(time.Second)
		log, _ := io.ReadAll(handler.commands[id].logWriter)
		assert.Assert(t, strings.Contains(string(log), "total"))
		t.Run("duplicate", func(t *testing.T) {
			err := client.StartCommand(context.Background(), id,
				&executor.StartCommandRequest{
					Command: []string{"ls", "-alh"},
					Dir:     dir,
				})
			assert.ErrorContains(t, err, "command already exists")
		})
	})
	t.Run("long-run", func(t *testing.T) {
		err := client.StartCommand(context.Background(), t.Name(),
			&executor.StartCommandRequest{
				Command: []string{"tail", "-f", "/dev/null"},
				Dir:     dir,
			})
		assert.NilError(t, err)
		time.Sleep(time.Millisecond * 10)
		_ = handler.commands[t.Name()].Process.Kill()
	})
	t.Run("invalid-command", func(t *testing.T) {
		err := client.StartCommand(context.Background(), t.Name(),
			&executor.StartCommandRequest{
				Command: []string{"non-exists"},
				Dir:     dir,
			})
		assert.ErrorContains(t, err, "not found")
	})
	t.Run("bad_request", func(t *testing.T) {
		err := client.StartCommand(context.Background(), t.Name(),
			&executor.StartCommandRequest{
				Command: []string{},
				Dir:     dir,
			})
		assert.ErrorContains(t, err, "command cannot be empty")
	})
}
