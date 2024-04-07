package handler

import (
	"context"
	"os"
	"runtime"

	"github.com/cox96de/runner/app/executor/executorpb"
)

var _ executorpb.ExecutorServer = &Handler{}

func (h *Handler) Environment(_ context.Context, _ *executorpb.EnvironmentRequest) (*executorpb.EnvironmentResponse, error) {
	return &executorpb.EnvironmentResponse{
		Environment: os.Environ(),
	}, nil
}

func (h *Handler) GetRuntimeInfo(context.Context, *executorpb.GetRuntimeInfoRequest) (*executorpb.GetRuntimeInfoResponse, error) {
	return &executorpb.GetRuntimeInfoResponse{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}, nil
}
