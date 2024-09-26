package handler

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/log"
	"github.com/samber/lo"
)

func (h *Handler) CancelJobExecution(ctx context.Context, request *api.CancelJobExecutionRequest) (*api.CancelJobExecutionResponse, error) {
	logger := log.ExtractLogger(ctx).WithFields(log.Fields{
		"job_execution_id": request.JobExecutionID,
	})
	logger.Infof("cancel job execution")
	_, err := h.UpdateJobExecution(log.WithLogger(ctx, logger), &api.UpdateJobExecutionRequest{
		JobExecutionID: request.JobExecutionID,
		Status:         lo.ToPtr(api.StatusCanceling),
	})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to update job execution")
	}
	return &api.CancelJobExecutionResponse{}, nil
}
