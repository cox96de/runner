package handler

import (
	"bytes"
	"context"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cox96de/runner/util"

	internalmodel "github.com/cox96de/runner/internal/model"

	"gotest.tools/v3/fs"

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

func TestHandler_getCommandLogHandler(t *testing.T) {
	testServer, _ := setupHandler(t)
	client := executor.NewClient(testServer.URL)
	t.Run("fast", func(t *testing.T) {
		logs := `go: downloading gotest.tools/v3 v3.4.0
go: downloading github.com/google/go-cmp v0.5.5
?   	github.com/cox96de/runner/app/executor	[no test files]
ok  	github.com/cox96de/runner/app/executor/handler	1.053s	coverage: 63.6% of statements in ./...
?   	github.com/cox96de/runner/cmd/executor	[no test files]
ok  	github.com/cox96de/runner/internal/executor	0.028s	coverage: 22.4% of statements in ./...
?   	github.com/cox96de/runner/internal/model	[no test files]
?   	github.com/cox96de/runner/util	[no test files]`
		testDir := fs.NewDir(t, "test", fs.WithFile("test.log", logs, fs.WithMode(0o644)))
		err := client.StartCommand(context.Background(), t.Name(), &executor.StartCommandRequest{
			Command: []string{"cat", testDir.Join("test.log")},
			Dir:     os.TempDir(),
		})
		assert.Assert(t, err)
		l := client.GetCommandLogs(context.Background(), t.Name())
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, l)
		assert.NilError(t, err)
		assert.DeepEqual(t, buf.String(), logs)
	})
	t.Run("slow", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip slow test on windows")
		}
		commands := []string{"docker run --rm golang:1.20 bash -c \"go env -w GOPROXY=https://goproxy.cn,direct && go install github.com/go-delve/delve/cmd/dlv@latest\""}
		commandsEnv := util.CompileUnixScript(commands)
		err := client.StartCommand(context.Background(), t.Name(), &executor.StartCommandRequest{
			Command: []string{"/bin/sh", "-c", "printf '%s' \"$COMMANDS\" | /bin/sh"},
			Env:     map[string]string{"COMMANDS": commandsEnv},
		})
		assert.Assert(t, err)
		l := client.GetCommandLogs(context.Background(), t.Name())
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, l)
		assert.NilError(t, err)
	})
}

func TestHandler_getCommandStatusHandler(t *testing.T) {
	testServer, _ := setupHandler(t)
	client := executor.NewClient(testServer.URL)
	t.Run("running", func(t *testing.T) {
		err := client.StartCommand(context.Background(), t.Name(), &executor.StartCommandRequest{
			Command: []string{"sleep", "1"},
			Dir:     os.TempDir(),
		})
		assert.NilError(t, err)
		response, err := client.GetCommandStatus(context.Background(), t.Name())
		assert.NilError(t, err)
		assert.DeepEqual(t, response, &internalmodel.GetCommandStatusResponse{
			ExitCode: 0,
			Exit:     false,
		})
		time.Sleep(time.Second * 2)
		response, err = client.GetCommandStatus(context.Background(), t.Name())
		assert.NilError(t, err)
		assert.DeepEqual(t, response, &internalmodel.GetCommandStatusResponse{
			ExitCode: 0,
			Exit:     true,
		})
	})
	t.Run("exit_code", func(t *testing.T) {
		err := client.StartCommand(context.Background(), t.Name(), &executor.StartCommandRequest{
			Command: []string{"bash", "-c", "exit 2"},
			Dir:     os.TempDir(),
		})
		assert.NilError(t, err)
		for i := 0; i < 10; i++ {
			response, err := client.GetCommandStatus(context.Background(), t.Name())
			assert.NilError(t, err)
			if !response.Exit {
				t.Logf("not exited yet, retrying")
				time.Sleep(time.Millisecond * 10)
				continue
			}
			assert.DeepEqual(t, response, &internalmodel.GetCommandStatusResponse{
				ExitCode: 2,
				Exit:     true,
				Error:    "exit status 2",
			})
		}
	})
	t.Run("not_found", func(t *testing.T) {
		_, err := client.GetCommandStatus(context.Background(), t.Name())
		assert.ErrorContains(t, err, "command not found")
	})
}
