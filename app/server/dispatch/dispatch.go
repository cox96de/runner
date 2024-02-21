package dispatch

import (
	"context"
	"encoding/json"

	"github.com/cox96de/runner/api"
	"github.com/samber/lo"

	"github.com/cox96de/runner/db"
	"github.com/pkg/errors"
)

type Service struct {
	dbClient *db.Client
}

func NewService(dbClient *db.Client) *Service {
	return &Service{dbClient: dbClient}
}

// Dispatch dispatches the jobs and updates the job executions.
// It pushes the jobs to the queue if all the dependencies are completed and success.
// TODO: proceed dispatch if one job is failed, all the dependent jobs should be skipped.
func (s *Service) Dispatch(ctx context.Context, jobs []*db.Job, executions []*db.JobExecution) error {
	d := &dispatcher{
		jobs: make(map[string]*dispatchJob),
	}
	updateJobExecutionOptions, err := d.Dispatch(jobs, executions)
	if err != nil {
		return errors.WithMessage(err, "failed to calculate jobs")
	}
	err = s.dbClient.Transaction(func(client *db.Client) error {
		for _, option := range updateJobExecutionOptions {
			if err := client.UpdateJobExecution(ctx, option); err != nil {
				return errors.WithMessage(err, "failed to update job executions")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// dispatcher is a helper to dispatch the jobs.
// It stores the jobs and their executions and make it easier to calculate the dispatching.
type dispatcher struct {
	jobs map[string]*dispatchJob
}

// dispatchJob uses to store the job and its execution.
type dispatchJob struct {
	job       *db.Job
	execution *db.JobExecution
}

func (d *dispatcher) Dispatch(jobs []*db.Job, executions []*db.JobExecution) ([]*db.UpdateJobExecutionOption, error) {
	executionsMapByJobID := lo.SliceToMap(executions, func(item *db.JobExecution) (int64, *db.JobExecution) {
		return item.JobID, item
	})
	for _, job := range jobs {
		execution, ok := executionsMapByJobID[job.ID]
		if !ok {
			return nil, errors.Errorf("job %s has no execution", job.Name)
		}
		d.jobs[job.Name] = &dispatchJob{
			job:       job,
			execution: execution,
		}
	}
	var result []*db.UpdateJobExecutionOption
	for _, job := range d.jobs {
		if job.execution.Status.IsCompleted() {
			continue
		}
		updateJobExecutionOption, err := d.isAllPreCompleted(job)
		if err != nil {
			// TODO: update the job to failed
			continue
		}
		if updateJobExecutionOption == nil {
			continue
		}
		result = append(result, updateJobExecutionOption)
	}
	return result, nil
}

func (d *dispatcher) isAllPreCompleted(job *dispatchJob) (*db.UpdateJobExecutionOption, error) {
	if len(job.job.DependsOn) == 0 {
		return &db.UpdateJobExecutionOption{
			ID:     job.execution.ID,
			Status: lo.ToPtr(api.StatusQueued),
		}, nil
	}
	var dependsOn []string
	err := json.Unmarshal(job.job.DependsOn, &dependsOn)
	if err != nil {
		return nil, err
	}
	for _, depend := range dependsOn {
		depJob, ok := d.jobs[depend]
		if !ok {
			return nil, errors.Errorf("job %s depends on %s, but not found", job.job.Name, depend)
		}
		if !depJob.execution.Status.IsCompleted() {
			return nil, nil
		}
		if depJob.execution.Status != api.StatusSucceeded {
			return &db.UpdateJobExecutionOption{
				ID:     job.execution.ID,
				Status: lo.ToPtr(api.StatusSkipped),
			}, nil
		}
	}
	return &db.UpdateJobExecutionOption{
		ID:     job.execution.ID,
		Status: lo.ToPtr(api.StatusQueued),
	}, nil
}
