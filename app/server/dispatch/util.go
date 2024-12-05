package dispatch

import (
	"context"

	"github.com/cox96de/runner/telemetry/trace"
	"go.opentelemetry.io/otel/attribute"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
)

// UpdateJobExecution updates job execution status and job queue status.
// It insert a new job queue if the job execution status is queued.
func (s *Service) UpdateJobExecution(ctx context.Context, client *db.Client, option *db.UpdateJobExecutionOption) error {
	// TODO: Add more attribute.
	ctx, span := trace.Start(ctx, "dispatch.update_job_execution", trace.WithAttributes(
		attribute.Int64("job_execution_id", option.ID),
	))
	defer span.End()
	var updatedJob *db.JobExecution
	err := client.Transaction(func(client *db.Client) error {
		if option.Status != nil {
			switch {
			case option.Status.IsCompleted():
				span.AddEvent("Status is completed")
				if err := client.DeleteJobQueueByJobExecutionID(ctx, option.ID); err != nil {
					return err
				}
			case *option.Status == api.StatusQueued:
				span.AddEvent("Status is queued")
				jobExecution, err := client.GetJobExecution(ctx, option.ID)
				if err != nil {
					return errors.WithMessage(err, "failed to get job execution")
				}
				job, err := client.GetJobByID(ctx, jobExecution.JobID)
				if err != nil {
					return errors.WithMessage(err, "failed to get job")
				}
				runsOn, err := db.UnmarshalRunsOn(job.RunsOn)
				if err != nil {
					return errors.WithMessage(err, "failed to unmarshal runs on")
				}
				_, err = client.CreateJobQueues(ctx, []*db.CreateJobQueueOption{
					{
						JobExecutionID: option.ID,
						Label:          runsOn.Label,
						Status:         *option.Status,
					},
				})
				if err != nil {
					return errors.WithMessage(err, "failed to create job queues")
				}
			default:
				if err := client.UpdateJobQueueStatus(ctx, option.ID, *option.Status); err != nil {
					return errors.WithMessage(err, "failed to update job queue status")
				}
			}
		}
		var err error
		updatedJob, err = client.UpdateJobExecution(ctx, option)
		if err != nil {
			return errors.WithMessage(err, "failed to update job execution")
		}
		if err != nil {
			return errors.WithMessage(err, "failed to send job execution event")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.eventhook.SendJobExecutionEvent(ctx, updatedJob)
}
