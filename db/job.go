package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cox96de/runner/api"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/samber/lo"

	"github.com/pkg/errors"
)

type Job struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID       int64     `gorm:"column:pipeline_id"`
	Name             string    `gorm:"column:name"`
	RunsOn           []byte    `gorm:"column:runs_on"`
	WorkingDirectory string    `gorm:"column:working_directory"`
	EnvVar           []byte    `gorm:"column:env_var"`
	DependsOn        []byte    `gorm:"column:depends_on"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *Job) TableName() string {
	return "job"
}

type CreateJobOption struct {
	PipelineID       int64
	Name             string
	RunsOn           *api.RunsOn
	WorkingDirectory string
	EnvVar           map[string]string
	DependsOn        []string
}

// CreateJobs creates new jobs.
func (c *Client) CreateJobs(ctx context.Context, options []*CreateJobOption) ([]*Job, error) {
	jobs := make([]*Job, 0, len(options))
	var err error
	for _, opt := range options {
		job := &Job{
			Name:             opt.Name,
			WorkingDirectory: opt.WorkingDirectory,
		}
		job.RunsOn, err = json.Marshal(opt.RunsOn)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal job.RunsOn")
		}
		job.EnvVar, err = json.Marshal(opt.EnvVar)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal job.EnvVar")
		}
		job.DependsOn, err = json.Marshal(opt.DependsOn)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal job.DependsOn")
		}
		jobs = append(jobs, job)
	}
	if err := c.conn.WithContext(ctx).Create(jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

// GetJobByID returns a job by its ID.
func (c *Client) GetJobByID(ctx context.Context, id int64) (*Job, error) {
	job := &Job{}
	if err := c.conn.WithContext(ctx).First(job, id).Error; err != nil {
		return nil, err
	}
	return job, nil
}

// PackJob packs a job into api.Job.
func PackJob(j *Job, executions []*JobExecution, steps []*Step, stepExecutions []*StepExecution) (*api.Job, error) {
	runsOn := &api.RunsOn{}
	err := json.Unmarshal(j.RunsOn, runsOn)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal job.RunsOn")
	}
	envVar := make(map[string]string)
	err = json.Unmarshal(j.EnvVar, &envVar)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal job.EnvVar")
	}
	dependsOn := make([]string, 0)
	err = json.Unmarshal(j.DependsOn, &dependsOn)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal job.DependsOn")
	}
	packSteps, err := packSteps(steps, stepExecutions)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to pack steps")
	}
	return &api.Job{
		ID:               j.ID,
		PipelineID:       j.PipelineID,
		Name:             j.Name,
		RunsOn:           runsOn,
		WorkingDirectory: j.WorkingDirectory,
		EnvVar:           envVar,
		Executions: lo.Map(executions, func(e *JobExecution, _ int) *api.JobExecution {
			return PackJobExecution(e, stepExecutions)
		}),
		Steps:     packSteps,
		DependsOn: dependsOn,
		CreatedAt: timestamppb.New(j.CreatedAt),
		UpdatedAt: timestamppb.New(j.UpdatedAt),
	}, nil
}

// PackJobExecution packs a job execution into api.JobExecution.
// If steps is nil, it will not pack steps.
func PackJobExecution(j *JobExecution, steps []*StepExecution) *api.JobExecution {
	stepExecutions := lo.Map(steps, func(e *StepExecution, _ int) *api.StepExecution {
		return PackStepExecution(e)
	})
	// TODO: sort step.
	return &api.JobExecution{
		ID:          j.ID,
		JobID:       j.JobID,
		Status:      j.Status,
		Steps:       stepExecutions,
		StartedAt:   api.ConvertTime(j.StartedAt),
		CompletedAt: api.ConvertTime(j.CompletedAt),
		CreatedAt:   api.ConvertTime(j.CreatedAt),
		UpdatedAt:   api.ConvertTime(j.UpdatedAt),
	}
}

type JobExecution struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	JobID       int64      `gorm:"column:job_id"`
	Status      api.Status `gorm:"column:status"`
	StartedAt   *time.Time `gorm:"column:started_at"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *JobExecution) TableName() string {
	return "job_execution"
}

type CreateJobExecutionOption struct {
	JobID  int64
	Status api.Status
}

// CreateJobExecutions creates new job executions.
func (c *Client) CreateJobExecutions(ctx context.Context, options []*CreateJobExecutionOption) ([]*JobExecution, error) {
	executions := make([]*JobExecution, 0, len(options))
	for _, option := range options {
		execution := &JobExecution{
			JobID:  option.JobID,
			Status: option.Status,
		}
		executions = append(executions, execution)
	}
	if err := c.conn.WithContext(ctx).Create(executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}

func (c *Client) GetJobExecution(ctx context.Context, id int64) (*JobExecution, error) {
	execution := &JobExecution{}
	if err := c.conn.WithContext(ctx).First(execution, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return execution, nil
}

type UpdateJobExecutionOption struct {
	ID          int64
	Status      *api.Status
	StartedAt   *time.Time
	CompletedAt *time.Time
}

func (c *Client) UpdateJobExecution(ctx context.Context, option *UpdateJobExecutionOption) error {
	updateField := map[string]interface{}{}
	if option.Status != nil {
		updateField["status"] = *option.Status
	}
	if option.StartedAt != nil {
		updateField["started_at"] = *option.StartedAt
	}
	if option.CompletedAt != nil {
		updateField["completed_at"] = *option.CompletedAt
	}
	return c.conn.WithContext(ctx).Model(&JobExecution{}).Where("id = ?", option.ID).Updates(updateField).Error
}

func (c *Client) GetQueuedJobExecutions(ctx context.Context, limit int) ([]*JobExecution, error) {
	executions := make([]*JobExecution, 0, limit)
	if err := c.conn.WithContext(ctx).Where("status = ? order by id", api.StatusQueued).Limit(limit).Find(&executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}
