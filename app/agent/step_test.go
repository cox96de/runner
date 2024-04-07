package agent

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/cox96de/runner/api"
	mockapi "github.com/cox96de/runner/api/mock"
	"github.com/cox96de/runner/engine/shell"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"gotest.tools/v3/assert"
)

func TestExecution_executeStep(t *testing.T) {
	logBuf := &bytes.Buffer{}
	client := mockapi.NewMockServerClient(gomock.NewController(t))
	client.EXPECT().UpdateStepExecution(gomock.Any(), gomock.All()).
		Return(&api.UpdateStepExecutionResponse{}, nil).AnyTimes()
	client.EXPECT().UploadLogLines(gomock.Any(), gomock.All()).
		DoAndReturn(func(ctx context.Context, req *api.UpdateLogLinesRequest, co ...grpc.CallOption) (*api.UpdateLogLinesResponse, error) {
			for _, l := range req.Lines {
				logBuf.WriteString(l.Output + "\n")
			}
			return &api.UpdateLogLinesResponse{}, nil
		}).AnyTimes()
	eng := shell.NewEngine()
	e := NewExecution(eng, &api.Job{
		Executions: []*api.JobExecution{
			{},
		},
	}, client)
	var err error
	e.runner, err = eng.CreateRunner(context.Background(), e, &api.Job{})
	assert.NilError(t, err)
	err = e.runner.Start(context.Background())
	assert.NilError(t, err)
	t.Run("unix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("windows")
		}
		err = e.executeStep(context.Background(), &api.Step{
			Name: "test",
			Executions: []*api.StepExecution{
				{},
			},
			Commands: []string{"echo hello"},
		})
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(logBuf.String(), "echo hello"), logBuf.String())
	})
	t.Run("windows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("not windows")
		}
		err = e.executeStep(context.Background(), &api.Step{
			Name: "test",
			Executions: []*api.StepExecution{
				{},
			},
			Commands: []string{"Write-Output hello"},
		})
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(logBuf.String(), "hello"), logBuf.String())
	})
}
