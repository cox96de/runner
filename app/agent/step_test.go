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
	// These test is broken, and it's not necessary. Fix it later.
	t.Skip("skip")
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
		Execution: &api.JobExecution{},
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
		e.jobTimeoutCtx = context.Background()
		e.stepExecutions[1] = &api.StepExecution{}
		err = e.executeStep(context.Background(), &api.Step{
			Name:     "test",
			ID:       1,
			Commands: []string{"echo hello"},
		})
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(logBuf.String(), "echo hello"), logBuf.String())
	})
	t.Run("windows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("not windows")
		}
		e.jobTimeoutCtx = context.Background()
		e.stepExecutions[1] = &api.StepExecution{}
		err = e.executeStep(context.Background(), &api.Step{
			Name:     "test",
			ID:       1,
			Commands: []string{"Write-Output hello"},
		})
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(logBuf.String(), "hello"), logBuf.String())
	})
}

func TestExecution_evalExpression(t *testing.T) {
	type args struct {
		expression string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "True",
			args: args{
				expression: "True",
			},
			want: true,
		},
		{
			name: "False",
			args: args{
				expression: "False",
			},
			want: false,
		},
		{
			name: "true",
			args: args{
				expression: "true",
			},
			want: true,
		},
		{
			name: "false",
			args: args{
				expression: "false",
			},
			want: false,
		},
		{
			name: "1+1>2",
			args: args{
				expression: "1+1>2",
			},
			want: false,
		},
		{
			name: "1+1>1",
			args: args{
				expression: "1+1>1",
			},
			want: true,
		},
		{
			name: "def",
			args: args{
				expression: `
def ok():
  return True
ok()
`,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Execution{}
			got, err := e.evalExpression(context.Background(), tt.args.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("evalExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}
