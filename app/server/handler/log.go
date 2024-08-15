package handler

import (
	"context"

	"github.com/cox96de/runner/api"
)

func (h *Handler) UploadLogLines(ctx context.Context, request *api.UpdateLogLinesRequest) (*api.UpdateLogLinesResponse, error) {
	// TODO: check if job & executions exists
	err := h.logService.Append(ctx, request.JobExecutionID, request.Name, request.Lines)
	if err != nil {
		return nil, err
	}
	return &api.UpdateLogLinesResponse{}, nil
}

func (h *Handler) GetLogLines(ctx context.Context, request *api.GetLogLinesRequest) (*api.GetLogLinesResponse, error) {
	// TODO: check if job & executions exists
	limit := int64(-1)
	if request.Limit != nil {
		limit = *request.Limit
	}
	logLines, err := h.logService.GetLogLines(ctx, request.JobExecutionID, request.Name, request.Offset,
		limit)
	if err != nil {
		return nil, err
	}
	return &api.GetLogLinesResponse{
		Lines: logLines,
	}, nil
}
