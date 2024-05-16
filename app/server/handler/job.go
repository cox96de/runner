package handler

import (
	"context"
	"time"

	"github.com/cox96de/runner/log"
	"github.com/samber/lo"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/db"
	"github.com/pkg/errors"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/lib"
)

func (h *Handler) UpdateJobExecution(ctx context.Context, request *api.UpdateJobExecutionRequest) (*api.UpdateJobExecutionResponse, error) {
	logger := log.ExtractLogger(ctx).WithFields(log.Fields{
		"job_id":           request.JobID,
		"job_execution_id": request.JobExecutionID,
	})
	lock, err := h.locker.Lock(ctx, lib.BuildJobExecutionLockKey(request.JobExecutionID), "update_job_execution", time.Second)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to lock job execution '%d'", request.JobExecutionID)
	}
	// May be delay and retry ?
	if !lock {
		return nil, errors.Errorf("job execution '%d' is locked", request.JobExecutionID)
	}
	defer func() {
		_, _ = h.locker.Unlock(ctx, lib.BuildJobExecutionLockKey(request.JobExecutionID))
	}()
	jobExecution, err := h.db.GetJobExecution(ctx, request.JobExecutionID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get job execution '%d'", request.JobExecutionID)
	}
	if request.Status != nil {
		logger.Infof("update job execution status from %s to '%s'", jobExecution.Status, *request.Status)
		if !dispatch.CheckJobStatus(jobExecution.Status, *request.Status) {
			return nil, errors.Errorf("invalid status transition from '%s' to '%s'", jobExecution.Status, *request.Status)
		}
		jobExecution.Status = *request.Status
		updateJobExecutionOption := &db.UpdateJobExecutionOption{
			ID:     jobExecution.JobID,
			Status: request.Status,
		}
		switch {
		case *request.Status == api.StatusPreparing:
			// TODO: add preparing at.
		case *request.Status == api.StatusRunning:
			updateJobExecutionOption.StartedAt = lo.ToPtr(time.Now())
		case (*request).Status.IsCompleted():
			updateJobExecutionOption.CompletedAt = lo.ToPtr(time.Now())
			if err := h.logService.Archive(ctx, jobExecution.JobID, jobExecution.ID); err != nil {
				logger.WithError(err).Error("failed to archive logs")
			}
		}
		err := h.db.UpdateJobExecution(ctx, updateJobExecutionOption)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to update job execution '%d'", request.JobExecutionID)
		}
	}
	return &api.UpdateJobExecutionResponse{
		JobExecution: db.PackJobExecution(jobExecution, nil),
	}, nil
}

func (h *Handler) ListJobExecutions(ctx context.Context, in *api.ListJobExecutionsRequest) (*api.ListJobExecutionsResponse, error) {
	logger := log.ExtractLogger(ctx)
	logger.Infof("list job executions for job '%d'", in.JobID)
	jobExecutions, err := h.db.GetJobExecutionsByJobID(ctx, in.JobID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to list job executions for job '%d'", in.JobID)
	}
	stepExecutionMap, err := h.db.GetStepExecutionsByJobExecutionIDs(ctx, lo.Map(jobExecutions, func(item *db.JobExecution, _ int) int64 {
		return item.ID
	}))
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to list step executions for job '%d'", in.JobID)
	}
	return &api.ListJobExecutionsResponse{
		Jobs: lo.Map(jobExecutions, func(item *db.JobExecution, _ int) *api.JobExecution {
			return db.PackJobExecution(item, stepExecutionMap[item.ID])
		}),
	}, nil
}
