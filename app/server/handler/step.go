package handler

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/log"
	"github.com/samber/lo"
)

func (h *Handler) UpdateStepExecution(ctx context.Context, request *api.UpdateStepExecutionRequest) (*api.UpdateStepExecutionResponse, error) {
	logger := log.ExtractLogger(ctx).WithFields(log.Fields{
		"step_execution_id": request.StepExecutionID,
	})
	stepExecution, err := h.db.GetStepExecution(ctx, request.StepExecutionID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get step execution '%d'", request.StepExecutionID)
	}
	if request.Status != nil {
		logger.Infof("update step execution status from %s to '%s'", stepExecution.Status, *request.Status)
		if !dispatch.CheckStepStatus(stepExecution.Status, *request.Status) {
			return nil, errors.Errorf("invalid status transition from '%s' to '%s'", stepExecution.Status, *request.Status)
		}
		stepExecution.Status = *request.Status
		updateStepExecutionOption := &db.UpdateStepExecutionOption{
			ID:       stepExecution.ID,
			Status:   request.Status,
			ExitCode: request.ExitCode,
		}
		switch {
		case *request.Status == api.StatusPreparing:
			// TODO: add preparing at.
		case *request.Status == api.StatusRunning:
			updateStepExecutionOption.StartedAt = lo.ToPtr(time.Now())
		case (*request).Status.IsCompleted():
			updateStepExecutionOption.CompletedAt = lo.ToPtr(time.Now())
		}
		updatedStep, err := h.db.UpdateStepExecution(ctx, updateStepExecutionOption)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to update step execution '%d'", request.StepExecutionID)
		}
		if err = h.eventhook.SendStepExecutionEvent(ctx, updatedStep); err != nil {
			return nil, errors.WithMessagef(err, "failed to send step execution event '%d'", request.StepExecutionID)
		}
	}
	return &api.UpdateStepExecutionResponse{
		StepExecution: db.PackStepExecution(stepExecution),
	}, nil
}

func (h *Handler) GetStepExecution(ctx context.Context, request *api.GetStepExecutionRequest) (*api.GetStepExecutionResponse, error) {
	stepExecution, err := h.db.GetStepExecution(ctx, request.StepExecutionID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get step execution '%d'", request.StepExecutionID)
	}
	return &api.GetStepExecutionResponse{
		StepExecution: db.PackStepExecution(stepExecution),
	}, nil
}
