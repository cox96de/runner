package handler

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cox96de/runner/util"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gotest.tools/v3/fs"

	"gotest.tools/v3/assert"
)

func TestHandler_StartCommand(t *testing.T) {
	handler, addr := setupHandler(t)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	client := executorpb.NewExecutorClient(conn)

	t.Run("shot-run", func(t *testing.T) {
		var (
			dir      string
			commands []string
		)
		if runtime.GOOS == "windows" {
			dir = "c:\\"
			commands = []string{"powershell", "-Command", "Write-Host hello"}
		} else {
			dir = "/tmp"
			commands = []string{"echo", "hello"}
		}
		resp, err := client.StartCommand(context.Background(),
			&executorpb.StartCommandRequest{
				Commands: commands,
				Dir:      dir,
			})
		assert.NilError(t, err)
		// TODO: wait for the command to finish
		time.Sleep(time.Second)
		commandID := resp.CommandID
		assert.Assert(t, resp.Status.Pid > 0)
		t.Run("get_log", func(t *testing.T) {
			// Wait for the command to finish to fetch full log.
			_, err := client.WaitCommand(context.Background(), &executorpb.WaitCommandRequest{
				CommandID: commandID,
				Timeout:   int64(time.Second * 2),
			})
			assert.NilError(t, err)
			commandLogResp, err := client.GetCommandLog(context.Background(), &executorpb.GetCommandLogRequest{
				CommandID: commandID,
			})
			assert.NilError(t, err)
			log, err := executorpb.ReadAllFromCommandLog(commandLogResp)
			assert.NilError(t, err)
			if runtime.GOOS == "windows" {
			} else {
				assert.Assert(t, strings.Contains(log, "hello"), log)
			}
		})
	})
	t.Run("long-run", func(t *testing.T) {
		var (
			dir      string
			commands []string
		)
		if runtime.GOOS == "windows" {
			dir = "c:\\"
			commands = []string{"powershell", "-Command", "ping 0.0.0.0 -t"}
		} else {
			dir = "/tmp"
			commands = []string{"tail", "-f", "/dev/null"}
		}
		resp, err := client.StartCommand(context.Background(),
			&executorpb.StartCommandRequest{
				Commands: commands,
				Dir:      dir,
			})
		assert.NilError(t, err)
		pid := resp.Status.Pid
		assert.Assert(t, pid > 0)
		time.Sleep(time.Millisecond * 10)
		_ = handler.commands[resp.CommandID].Process.Kill()
	})
	t.Run("invalid-command", func(t *testing.T) {
		_, err := client.StartCommand(context.Background(),
			&executorpb.StartCommandRequest{
				Commands: []string{"non-exists"},
				Dir:      "/tmp",
			})
		assert.ErrorContains(t, err, "not found")
	})
	t.Run("emtpy_command", func(t *testing.T) {
		_, err := client.StartCommand(context.Background(),
			&executorpb.StartCommandRequest{
				Commands: []string{},
				Dir:      "/tmp",
			})
		assert.ErrorContains(t, err, "no command provided")
	})
}

func TestHandler_GetCommandLog(t *testing.T) {
	_, addr := setupHandler(t)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	client := executorpb.NewExecutorClient(conn)
	t.Run("fast", func(t *testing.T) {
		logs := `go: downloading gotest.tools/v3 v3.4.0
go: downloading github.com/google/go-cmp v0.5.5
?   	github.com/cox96de/runner/app/executor	[no test files]
ok  	github.com/cox96de/runner/app/executor/handler	1.053s	coverage: 63.6% of statements in ./...
?   	github.com/cox96de/runner/cmd/executor	[no test files]
ok  	github.com/cox96de/runner/internal/executor	0.028s	coverage: 22.4% of statements in ./...
?   	github.com/cox96de/runner/internal/model	[no test files]
?   	github.com/cox96de/runner/util	[no test files]
`
		testDir := fs.NewDir(t, "test", fs.WithFile("test.log", logs, fs.WithMode(0o644)))
		var (
			dir      string
			commands []string
		)
		if runtime.GOOS == "windows" {
			dir = "c:\\"
			commands = []string{"powershell", "-Command", "Get-Content", "-Path", testDir.Join("test.log")}
		} else {
			dir = "/tmp"
			commands = []string{"cat", testDir.Join("test.log")}
		}
		resp, err := client.StartCommand(context.Background(), &executorpb.StartCommandRequest{
			Commands: commands,
			Dir:      dir,
		})
		assert.Assert(t, err)
		getLogResp, err := client.GetCommandLog(context.Background(), &executorpb.GetCommandLogRequest{
			CommandID: resp.CommandID,
		})
		assert.NilError(t, err)
		log, err := executorpb.ReadAllFromCommandLog(getLogResp)
		assert.NilError(t, err)
		if runtime.GOOS == "windows" {
			log = strings.Replace(log, "\r\n", "\n", -1)
		}
		assert.DeepEqual(t, log, logs)
	})
	t.Run("slow", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip slow test on windows")
		}
		commands := []string{"echo 1", "sleep 1", "echo 2"}
		commandsEnv := util.CompileUnixScript(commands)
		resp, err := client.StartCommand(context.Background(), &executorpb.StartCommandRequest{
			Commands: []string{"/bin/sh", "-c", "printf '%s' \"$COMMANDS\" | /bin/sh"},
			Env:      []string{"COMMANDS=" + commandsEnv},
		})
		assert.Assert(t, err)
		getLogResp, err := client.GetCommandLog(context.Background(), &executorpb.GetCommandLogRequest{
			CommandID: resp.CommandID,
		})
		assert.NilError(t, err)
		log, err := executorpb.ReadAllFromCommandLog(getLogResp)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(log, "echo 1"))
		assert.Assert(t, strings.Contains(log, "echo 2"))
	})
}

func TestHandler_WaitCommand(t *testing.T) {
	_, addr := setupHandler(t)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	client := executorpb.NewExecutorClient(conn)
	t.Run("wait", func(t *testing.T) {
		var commands []string
		if runtime.GOOS == "windows" {
			commands = []string{"powershell", "-Command", "sleep 1"}
		} else {
			commands = []string{"sleep", "1"}
		}
		resp, err := client.StartCommand(context.Background(), &executorpb.StartCommandRequest{
			Commands: commands,
			Dir:      os.TempDir(),
		})
		assert.NilError(t, err)
		response, err := client.WaitCommand(context.Background(), &executorpb.WaitCommandRequest{
			CommandID: resp.CommandID,
			Timeout:   int64(time.Millisecond * 10),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, response.Status.Exit, false)
		response, err = client.WaitCommand(context.Background(), &executorpb.WaitCommandRequest{
			CommandID: resp.CommandID,
			Timeout:   int64(time.Second * 5),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, response.Status.Exit, true)
	})
	t.Run("exit_code", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip on windows")
		}
		resp, err := client.StartCommand(context.Background(), &executorpb.StartCommandRequest{
			Commands: []string{"bash", "-c", "exit 2"},
			Dir:      os.TempDir(),
		})
		assert.NilError(t, err)
		response, err := client.WaitCommand(context.Background(), &executorpb.WaitCommandRequest{
			CommandID: resp.CommandID,
			Timeout:   int64(time.Second * 2),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, response.Status, &executorpb.ProcessStatus{
			Pid:      resp.Status.Pid,
			ExitCode: 2,
			Exit:     true,
			Error:    "exit status 2",
		}, cmpopts.IgnoreUnexported(executorpb.ProcessStatus{}))
	})
}

func TestSetCommandMockRandomString(t *testing.T) {

	// Execute normally.
	t.Run("normal_execute", func(t *testing.T) {

		h := &Handler{
			commands: make(map[string]*command),
		}

		originalRandomStringFunc := randomStringFunc
		defer func() {
			randomStringFunc = originalRandomStringFunc
		}()

		commandID, _ := h.setCommand(&command{})
		_, exists := h.commands[commandID]
		assert.Assert(t, exists)
	})

	// Successfully set commandID after conflict occurs.
	t.Run("set_commandID_after_conflict", func(t *testing.T) {

		h := &Handler{
			commands: make(map[string]*command),
		}

		originalRandomStringFunc := randomStringFunc
		defer func() {
			randomStringFunc = originalRandomStringFunc
		}()

		callCount := 0
		mock_commandID := []int{0, 0, 0, 0, 2, 3}
		randomStringFunc = func(length int) string {
			callCount++
			return fmt.Sprintf("mockID%d", mock_commandID[callCount])
		}
		commandID1, err1 := h.setCommand(&command{})
		commandID_conflit, err2 := h.setCommand(&command{})

		// _, exists_1 := h.commands[commandID1]
		// _, exists_2 := h.commands[commandID_conflit]

		assert.Assert(t, commandID1 == "mockID0")
		assert.Assert(t, commandID_conflit == "mockID2")
		assert.NilError(t, err1)
		assert.NilError(t, err2)
	})

	// After 10 conflicts, commandID failed to be set.
	t.Run("fail_to_set_commandID", func(t *testing.T) {

		h := &Handler{
			commands: make(map[string]*command),
		}

		originalRandomStringFunc := randomStringFunc
		defer func() {
			randomStringFunc = originalRandomStringFunc
		}()

		callCount := 0
		mock_commandID := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		randomStringFunc = func(length int) string {
			callCount++
			return fmt.Sprintf("mockID%d", mock_commandID[callCount])
		}
		commandID1, err1 := h.setCommand(&command{})
		commandID_conflit, err2 := h.setCommand(&command{})

		// _, exists_1 := h.commands[commandID1]
		// _, exists_2 := h.commands[commandID_conflit]

		assert.Assert(t, commandID1 == "mockID0")
		assert.Assert(t, commandID_conflit == "")
		assert.NilError(t, err1)
		assert.Error(t, err2, "can not get a valid commandID")
	})
}
