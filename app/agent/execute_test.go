package agent

import (
	"context"
	"runtime"
	"testing"

	"google.golang.org/grpc"

	mockapi "github.com/cox96de/runner/api/mock"
	"go.uber.org/mock/gomock"

	"github.com/cox96de/runner/api"

	log "github.com/sirupsen/logrus"

	"github.com/cox96de/runner/engine/shell"
	"github.com/cox96de/runner/testtool"

	"gotest.tools/v3/assert"
)

func TestExecutor_executeJob(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip test on windows")
	}
	e := shell.NewEngine()
	log.SetLevel(log.DebugLevel)
	gitRoot, err := testtool.GetRepositoryRoot()
	assert.NilError(t, err)
	job := &api.Job{
		Steps: []*api.Step{
			{
				Name:             "step1",
				Commands:         []string{"ls -alh"},
				WorkingDirectory: gitRoot,
				Executions: []*api.StepExecution{
					{
						ID: 1,
					},
				},
			},
			{
				Name:             "step2",
				Commands:         []string{"pwd"},
				WorkingDirectory: gitRoot,
				Executions: []*api.StepExecution{
					{
						ID: 2,
					},
				},
			},
		},
		Executions: []*api.JobExecution{
			{
				Steps: []*api.StepExecution{
					{
						ID:             1,
						JobExecutionID: 1,
					},
				},
			},
		},
	}
	client := mockapi.NewMockServerClient(gomock.NewController(t))
	client.EXPECT().UpdateJobExecution(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
		request *api.UpdateJobExecutionRequest, option ...grpc.CallOption,
	) (*api.UpdateJobExecutionResponse, error) {
		return &api.UpdateJobExecutionResponse{Job: &api.JobExecution{
			Status: *request.Status,
		}}, nil
	}).AnyTimes()
	client.EXPECT().UpdateStepExecution(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
		request *api.UpdateStepExecutionRequest, option ...grpc.CallOption,
	) (*api.UpdateStepExecutionResponse, error) {
		return &api.UpdateStepExecutionResponse{Step: &api.StepExecution{
			Status: *request.Status,
		}}, nil
	}).AnyTimes()
	client.EXPECT().UploadLogLines(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&api.UpdateLogLinesResponse{}, nil).AnyTimes()
	execution := NewExecution(e, job, client)
	err = execution.Execute(context.Background())
	assert.NilError(t, err)
}
