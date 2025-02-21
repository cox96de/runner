package handler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/lib"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gotest.tools/v3/fs"

	"gotest.tools/v3/assert"
)

func TestHandler_StartCommand(t *testing.T) {
	handler, addr := setupHandler(t)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		_, err = client.WaitCommand(context.Background(), &executorpb.WaitCommandRequest{
			CommandID: resp.CommandID,
		})
		assert.NilError(t, err)
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
	t.Run("empty_command", func(t *testing.T) {
		_, err := client.StartCommand(context.Background(),
			&executorpb.StartCommandRequest{
				Commands: []string{},
				Dir:      "/tmp",
			})
		assert.ErrorContains(t, err, "no command provided")
	})
	t.Run("workdir", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip on windows")
		}
		t.Run("with_dir", func(t *testing.T) {
			dir := fs.NewDir(t, "test")
			target := filepath.Join(dir.Path(), "to", "be", "created")
			_, err := client.StartCommand(context.Background(),
				&executorpb.StartCommandRequest{
					Commands: []string{"ls", "-alh"},
					Dir:      target,
				})
			assert.NilError(t, err)
			stat, err := os.Stat(target)
			assert.NilError(t, err)
			assert.Assert(t, stat.IsDir())
		})
		t.Run("workdir_is_file", func(t *testing.T) {
			dir := fs.NewDir(t, "test", fs.WithFile("file", "content"))
			target := filepath.Join(dir.Path(), "file")
			_, err := client.StartCommand(context.Background(),
				&executorpb.StartCommandRequest{
					Commands: []string{"ls", "-alh"},
					Dir:      target,
				})
			assert.ErrorContains(t, err, "is not a directory")
		})
	})
	t.Run("username", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip on windows")
		}
		t.Run("non_exists", func(t *testing.T) {
			_, err := client.StartCommand(context.Background(),
				&executorpb.StartCommandRequest{
					Commands: []string{"ls", "-alh"},
					Username: "nonexists",
				})
			assert.ErrorContains(t, err, "failed to find user")
		})
		t.Run("with_username", func(t *testing.T) {
			t.Skip("don't support in github action")
			if runtime.GOOS == "windows" {
				t.Skip("skip on windows")
			}
			username := "ci_test"
			_, err := user.Lookup(username)

			var userError user.UnknownUserError
			if err != nil && errors.As(err, &userError) {
				output, err := Run("useradd", username)
				assert.NilError(t, err, output)
				t.Cleanup(func() {
					_, _ = Run("userdel", username)
				})
			}
			startCommandResp, err := client.StartCommand(context.Background(),
				&executorpb.StartCommandRequest{
					Commands: []string{"whoami"},
					Username: username,
				})
			assert.NilError(t, err)
			commandLogResp, err := client.GetCommandLog(context.Background(), &executorpb.GetCommandLogRequest{
				CommandID: startCommandResp.CommandID,
			})
			assert.NilError(t, err)
			log, err := executorpb.ReadAllFromCommandLog(commandLogResp)
			assert.NilError(t, err)
			assert.Equal(t, log, username+"\n")
		})
	})
}

func TestHandler_GetCommandLog(t *testing.T) {
	_, addr := setupHandler(t)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	client := executorpb.NewExecutorClient(conn)
	t.Run("fast", func(t *testing.T) {
		expectLogs := `go: downloading gotest.tools/v3 v3.4.0
go: downloading github.com/google/go-cmp v0.5.5
?   	github.com/cox96de/runner/app/executor	[no test files]
ok  	github.com/cox96de/runner/app/executor/handler	1.053s	coverage: 63.6% of statements in ./...
?   	github.com/cox96de/runner/cmd/executor	[no test files]
ok  	github.com/cox96de/runner/internal/executor	0.028s	coverage: 22.4% of statements in ./...
?   	github.com/cox96de/runner/internal/model	[no test files]
?   	github.com/cox96de/runner/util	[no test files]
`
		testDir := fs.NewDir(t, "test", fs.WithFile("test.log", expectLogs, fs.WithMode(0o644)))
		var (
			dir      string
			commands []string
		)
		if runtime.GOOS == "windows" {
			dir = "c:\\"
			commands = []string{"Get-Content -Path " + testDir.Join("test.log")}
		} else {
			dir = "/tmp"
			commands = []string{"cat " + testDir.Join("test.log")}
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
		assert.Assert(t, strings.Contains(log, expectLogs), log)
	})
	t.Run("slow", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip slow test on windows")
		}
		commands := []string{"echo 1", "sleep 1", "echo 2"}
		resp, err := client.StartCommand(context.Background(), &executorpb.StartCommandRequest{
			Commands: commands,
			Env:      os.Environ(),
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
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	client := executorpb.NewExecutorClient(conn)
	t.Run("wait", func(t *testing.T) {
		commands := []string{"sleep 1"}
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
			Timeout:   int64(time.Second * 10),
		})
		assert.NilError(t, err)
		assert.Equal(t, response.Status.Exit, true, "pid: %+v", response.Status)
	})
	t.Run("exit_code", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skip on windows")
		}
		resp, err := client.StartCommand(context.Background(), &executorpb.StartCommandRequest{
			Commands: []string{"exit 2"},
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
		mockCommandID := []int{0, 0, 0, 0, 2, 3}
		randomStringFunc = func(length int) string {
			callCount++
			return fmt.Sprintf("mockID%d", mockCommandID[callCount])
		}
		commandID1, err1 := h.setCommand(&command{})
		commandIDConflit, err2 := h.setCommand(&command{})

		assert.Assert(t, commandID1 == "mockID0")
		assert.Assert(t, commandIDConflit == "mockID2")
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

		randomStringFunc = func(length int) string {
			return "mockID0"
		}
		commandID1, err1 := h.setCommand(&command{})
		commandIDConflit, err2 := h.setCommand(&command{})

		assert.Assert(t, commandID1 == "mockID0")
		assert.Assert(t, commandIDConflit == "")
		assert.NilError(t, err1)
		assert.Error(t, err2, "can not get a valid commandID")
	})
}

func Test_newCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows")
	}
	rb := lib.NewRingBuffer(defaultRingBufferSize)
	c, err := newCommand([]string{"python3", "-m", "non-exists"}, rb, rb, "", os.Environ(), "")
	assert.NilError(t, err)
	err = c.Start()
	assert.NilError(t, err)
	c.Wait()
}

// Run starts a command and waits.
// The output includes its stdout & stderr. err is not nil when an error occurred or the exit code is not 0.
func Run(command string, args ...string) (output string, err error) {
	cmd := exec.Command(command, args...)
	ouptut, err := cmd.CombinedOutput()
	if err != nil {
		return string(ouptut), errors.WithMessage(err, "failed to run command")
	}
	if exitCode := cmd.ProcessState.ExitCode(); exitCode != 0 {
		return string(ouptut), errors.Errorf("command exited with code %d", exitCode)
	}
	return string(ouptut), nil
}
