package handler

import (
	"context"
	"time"

	"github.com/cox96de/runner/telemetry/trace"
	"go.opentelemetry.io/otel/attribute"

	"github.com/cox96de/runner/log"
	"github.com/samber/lo"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/db"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/lib"
)

func (h *Handler) UpdateJobExecution(ctx context.Context, request *api.UpdateJobExecutionRequest) (*api.UpdateJobExecutionResponse, error) {
	logger := log.ExtractLogger(ctx).WithFields(log.Fields{
		"job_execution_id": request.JobExecutionID,
	})
	ctx, span := trace.Start(ctx, "handler.update_job_execution",
		trace.WithAttributes(attribute.Int64("job_execution_id", request.JobExecutionID)))
	defer span.End()
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
	targetStatus := request.Status
	if targetStatus != nil {
		// If not dispatch, cancel would lead to failed.
		if *targetStatus == api.StatusCanceling && jobExecution.Status.IsPreDispatch() {
			logger.Infof("job is not dispatched, cancel directly")
			targetStatus = lo.ToPtr(api.StatusFailed)
			request.Reason = &api.Reason{Reason: api.FailedReasonCancelled}
		}
		logger.Infof("update job execution status from %s to '%s'", jobExecution.Status, *targetStatus)
		if !dispatch.CheckJobStatus(jobExecution.Status, *targetStatus) {
			return nil, errors.Errorf("invalid status transition from '%s' to '%s'", jobExecution.Status, *targetStatus)
		}

		jobExecution.Status = *targetStatus
		updateJobExecutionOption := &db.UpdateJobExecutionOption{
			ID:     jobExecution.ID,
			Status: targetStatus,
			Reason: request.Reason,
		}
		switch {
		case *targetStatus == api.StatusPreparing:
			// TODO: add preparing at.
		case *targetStatus == api.StatusRunning:
			updateJobExecutionOption.StartedAt = lo.ToPtr(time.Now())
		case targetStatus.IsCompleted():
			// TODO: assign completed at in dispatch.UpdateJobExecution
			updateJobExecutionOption.CompletedAt = lo.ToPtr(time.Now())
			if err := h.logService.Archive(ctx, jobExecution.ID); err != nil {
				logger.WithError(err).Error("failed to archive logs")
			}
		}
		err := h.dispatchService.UpdateJobExecution(ctx, h.db, updateJobExecutionOption)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to update job execution '%d'", request.JobExecutionID)
		}
	}
	packJobExecution, err := db.PackJobExecution(jobExecution, nil)
	return &api.UpdateJobExecutionResponse{
		JobExecution: packJobExecution,
	}, err
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

	resp := &api.ListJobExecutionsResponse{}
	for _, execution := range jobExecutions {
		e, err := db.PackJobExecution(execution, stepExecutionMap[execution.ID])
		if err != nil {
			return nil, errors.WithMessage(err, "failed to pack job execution")
		}
		resp.Jobs = append(resp.Jobs, e)
	}
	return resp, nil
}

func (h *Handler) GetJobExecution(ctx context.Context, in *api.GetJobExecutionRequest) (*api.GetJobExecutionResponse, error) {
	logger := log.ExtractLogger(ctx)
	logger.Infof("get job executions for job exeuction id: %d", in.JobExecutionID)
	jobExecutionPO, err := h.db.GetJobExecution(ctx, in.JobExecutionID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get job execution '%d'", in.JobExecutionID)
	}
	var steps []*db.StepExecution
	if in.WithStepExecution != nil && *in.WithStepExecution {
		steps, err = h.db.GetStepExecutionsByJobExecutionID(ctx, in.JobExecutionID)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to get step executions")
		}
	}
	jobExecution, err := db.PackJobExecution(jobExecutionPO, steps)
	return &api.GetJobExecutionResponse{JobExecution: jobExecution}, err
}

// RerunJob rerun the job by the latest job execution.
// TODO: this implementation is not correct, rerun should create new job executions include jobs depends on it.
func (h *Handler) RerunJob(ctx context.Context, in *api.RerunJobRequest) (*api.RerunJobResponse, error) {
	logger := log.ExtractLogger(ctx)
	logger.Infof("rerun job: %d", in.JobID)
	var latestJobExecution *db.JobExecution
	jobExecutions, err := h.db.GetJobExecutionsByJobID(ctx, in.JobID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to job executions '%d'", in.JobID)
	}
	// Get the latest job execution.
	latestJobExecution = getLatestJobExecution(jobExecutions)
	if !latestJobExecution.Status.IsCompleted() {
		return nil, errors.Errorf("job '%d' is not completed", in.JobID)
	}
	job, err := h.db.GetJobByID(ctx, latestJobExecution.JobID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get job '%d'", latestJobExecution.JobID)
	}
	latestJobExecution, err = h.resetJobExecution(ctx, latestJobExecution)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to reset job execution")
	}
	// FIXME: if dispatch failed, cannot redispatch.
	err = h.dispatchService.Dispatch(ctx, []*db.Job{job}, []*db.JobExecution{latestJobExecution})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to dispatch job")
	}
	execution, err := db.PackJobExecution(latestJobExecution, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to pack job execution")
	}
	// FIXME: the status of job execution might be not correct because of dispatching.
	return &api.RerunJobResponse{
		JobExecution: execution,
	}, nil
}

func (h *Handler) resetJobExecution(ctx context.Context, latestJobExecution *db.JobExecution) (*db.JobExecution, error) {
	err := h.db.Transaction(func(dbClient *db.Client) error {
		stepExecutions, err := dbClient.GetStepExecutionsByJobExecutionID(ctx, latestJobExecution.ID)
		if err != nil {
			return errors.WithMessagef(err, "failed to get step executions for job execution '%d'", latestJobExecution.ID)
		}
		err = dbClient.ResetStepExecutions(ctx, lo.Map(stepExecutions, func(stepExecution *db.StepExecution, _ int) int64 {
			return stepExecution.ID
		}))
		if err != nil {
			return errors.WithMessage(err, "failed to update step executions")
		}
		err = dbClient.ResetJobExecution(ctx, latestJobExecution.ID)
		if err != nil {
			return errors.WithMessagef(err, "failed to reset job execution '%d'", latestJobExecution.ID)
		}
		latestJobExecution, err = dbClient.GetJobExecution(ctx, latestJobExecution.ID)
		if err != nil {
			return errors.WithMessagef(err, "failed to get job execution '%d'", latestJobExecution.ID)
		}
		return nil
	})
	return latestJobExecution, err
}

func getLatestJobExecution(jobExecutions []*db.JobExecution) *db.JobExecution {
	var latestJobExecution *db.JobExecution
	for _, jobExecution := range jobExecutions {
		if latestJobExecution == nil || jobExecution.CreatedAt.After(latestJobExecution.CreatedAt) {
			latestJobExecution = jobExecution
		}
	}
	return latestJobExecution
}
