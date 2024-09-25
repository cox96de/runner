package monitor

import (
	"context"
	"time"

	"github.com/cox96de/runner/telemetry/trace"
	"go.opentelemetry.io/otel/attribute"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/log"
	"github.com/samber/lo"
)

// RecycleHeartbeatTimeoutJobs recycle heartbeat timeout jobs.
// If the job execution is not updated for a long time, it will be recycled (update status to failed).
func (s *Service) RecycleHeartbeatTimeoutJobs(ctx context.Context, timeout time.Duration) error {
	jobQueues, err := s.db.ListHeartbeatJobExecutions(ctx, timeout)
	if err != nil {
		return errors.WithMessage(err, "failed to list heartbeat timeout job executions")
	}
	for _, jobQueue := range jobQueues {
		if !jobQueue.Status.IsRunning() {
			continue
		}
		if err := s.recycleHeartbeatTimeoutJob(ctx, jobQueue.JobExecutionID); err != nil {
			log.ExtractLogger(ctx).Errorf("failed to recycle heartbeat timeout: %+v", err)
		}
	}
	return nil
}

func (s *Service) recycleHeartbeatTimeoutJob(ctx context.Context, jobExecutionID int64) error {
	logger := log.ExtractLogger(ctx)
	logger.Infof("recycle job execution: %d", jobExecutionID)
	ctx, span := trace.Start(ctx, "cronjob.recycle_heartbeat_timeout_job",
		trace.WithAttributes(attribute.Int64("job_execution_id", jobExecutionID)))
	defer span.End()
	span.AddEvent("Recycle heartbeat timeout job")
	err := s.db.Transaction(func(client *db.Client) error {
		if err := completedUnfinishedSteps(ctx, client, jobExecutionID); err != nil {
			return errors.WithMessagef(err, "failed to complete unfinished steps of job execution %d", jobExecutionID)
		}
		return dispatch.UpdateJobExecution(ctx, client, &db.UpdateJobExecutionOption{
			ID:     jobExecutionID,
			Status: lo.ToPtr(api.StatusFailed),
			Reason: &api.Reason{
				Reason:  api.FailedReasonHeartbeatTimeout,
				Message: "",
			},
			CompletedAt: lo.ToPtr(time.Now()),
		})
	})
	if err != nil {
		return errors.WithMessage(err, "failed to recycle heartbeat timeout job")
	}
	return s.logstorageService.Archive(ctx, jobExecutionID)
}

func completedUnfinishedSteps(ctx context.Context, dbCli *db.Client, jobExecutionID int64) error {
	steps, err := dbCli.GetStepExecutionsByJobExecutionID(ctx, jobExecutionID)
	if err != nil {
		return errors.WithMessage(err, "failed to get step executions")
	}
	for _, step := range steps {
		if step.Status.IsCompleted() {
			continue
		}
		err := dbCli.UpdateStepExecution(ctx, &db.UpdateStepExecutionOption{
			ID:     step.ID,
			Status: lo.ToPtr(api.StatusSkipped),
		})
		if err != nil {
			return errors.WithMessagef(err, "failed to update step execution %d", step.ID)
		}
	}
	return nil
}
