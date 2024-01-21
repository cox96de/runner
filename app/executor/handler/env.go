package handler

import (
	"context"
	"os"

	"github.com/cox96de/runner/app/executor/executorpb"
)

var _ executorpb.ExecutorServer = &Handler{}

func (h *Handler) Environment(_ context.Context, _ *executorpb.EnvironmentRequest) (*executorpb.EnvironmentResponse, error) {
	return &executorpb.EnvironmentResponse{
		Environment: os.Environ(),
	}, nil
}
