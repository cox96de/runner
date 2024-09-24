package handler

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
)

func (h *Handler) Heartbeat(ctx context.Context, request *api.HeartbeatRequest) (*api.HeartbeatResponse, error) {
	jobExecution, err := h.db.GetJobExecution(ctx, request.JobExecutionID)
	if err != nil {
		if db.IsRecordNotFoundError(err) {
			return nil, &HTTPError{
				Code:       http.StatusMethodNotAllowed,
				CauseError: errors.New("Job is not running"),
			}
		}
		return nil, errors.WithMessagef(err, "failed to get job execution by id %d", request.JobExecutionID)
	}
	if err = h.db.TouchHeartbeat(ctx, jobExecution.ID); err != nil {
		return nil, errors.WithMessagef(err, "failed to touch heartbeat for job execution %d", jobExecution.ID)
	}
	return &api.HeartbeatResponse{}, nil
}
