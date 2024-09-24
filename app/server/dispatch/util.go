package dispatch

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
)

// UpdateJobExecution updates job execution status and job queue status.
// It insert a new job queue if the job execution status is queued.
func UpdateJobExecution(ctx context.Context, client *db.Client, option *db.UpdateJobExecutionOption) error {
	return client.Transaction(func(client *db.Client) error {
		if option.Status != nil {
			switch {
			case option.Status.IsCompleted():
				if err := client.DeleteJobQueueByJobExecutionID(ctx, option.ID); err != nil {
					return err
				}
			case *option.Status == api.StatusQueued:
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
		err := client.UpdateJobExecution(ctx, option)
		if err != nil {
			return errors.WithMessage(err, "failed to update job execution")
		}
		return err
	})
}
