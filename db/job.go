package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cox96de/runner/api"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/samber/lo"

	"github.com/cockroachdb/errors"
)

type Job struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID       int64     `gorm:"column:pipeline_id"`
	Name             string    `gorm:"column:name"`
	RunsOn           []byte    `gorm:"column:runs_on"`
	WorkingDirectory string    `gorm:"column:working_directory"`
	EnvVar           []byte    `gorm:"column:env_var"`
	DependsOn        []byte    `gorm:"column:depends_on"`
	Timeout          int32     `gorm:"column:timeout"`
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
	Timeout          int32
}

// CreateJobs creates new jobs.
func (c *Client) CreateJobs(ctx context.Context, options []*CreateJobOption) ([]*Job, error) {
	jobs := make([]*Job, 0, len(options))
	var err error
	for _, opt := range options {
		job := &Job{
			Name:             opt.Name,
			PipelineID:       opt.PipelineID,
			WorkingDirectory: opt.WorkingDirectory,
			Timeout:          opt.Timeout,
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
// latestExecution is the latest job execution.
// executions are all job executions.
// steps are all steps of the job.
// stepExecutions are all step executions of the job.
func PackJob(j *Job, latestExecution *JobExecution, executions []*JobExecution, steps []*Step,
	stepExecutions []*StepExecution,
) (*api.Job, error) {
	executionsByJobExecutionID := lo.GroupBy(stepExecutions, func(item *StepExecution) int64 {
		return item.JobExecutionID
	})
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
	var jobExecutions []*api.JobExecution
	for _, e := range executions {
		je, err := PackJobExecution(e, stepExecutions)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to pack job execution")
		}
		jobExecutions = append(jobExecutions, je)
	}
	latestExe, err := PackJobExecution(latestExecution, executionsByJobExecutionID[latestExecution.ID])
	if err != nil {
		return nil, errors.WithMessage(err, "failed to pack latest job execution")
	}
	p := &api.Job{
		ID:               j.ID,
		PipelineID:       j.PipelineID,
		Name:             j.Name,
		RunsOn:           runsOn,
		WorkingDirectory: j.WorkingDirectory,
		EnvVar:           envVar,
		Executions:       jobExecutions,
		Execution:        latestExe,
		Steps:            packSteps,
		DependsOn:        dependsOn,
		Timeout:          j.Timeout,
		CreatedAt:        timestamppb.New(j.CreatedAt),
		UpdatedAt:        timestamppb.New(j.UpdatedAt),
	}
	return p, nil
}

func UnmarshalRunsOn(body []byte) (*api.RunsOn, error) {
	r := &api.RunsOn{}
	err := json.Unmarshal(body, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// PackJobExecution packs a job execution into api.JobExecution.
// If steps is nil, it will not pack steps.
func PackJobExecution(j *JobExecution, steps []*StepExecution) (*api.JobExecution, error) {
	if j == nil {
		return nil, nil
	}
	stepExecutions := lo.Map(steps, func(e *StepExecution, _ int) *api.StepExecution {
		return PackStepExecution(e)
	})
	// TODO: sort step.
	p := &api.JobExecution{
		ID:          j.ID,
		JobID:       j.JobID,
		Status:      j.Status,
		Steps:       stepExecutions,
		StartedAt:   api.ConvertTime(j.StartedAt),
		CompletedAt: api.ConvertTime(j.CompletedAt),
		CreatedAt:   api.ConvertTime(j.CreatedAt),
		UpdatedAt:   api.ConvertTime(j.UpdatedAt),
	}
	if len(j.Reason) > 0 {
		if err := json.Unmarshal(j.Reason, &p.Reason); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal job execution reason")
		}
	}
	return p, nil
}

type JobExecution struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	JobID       int64      `gorm:"column:job_id"`
	Status      api.Status `gorm:"column:status"`
	Reason      []byte     `gorm:"column:reason"`
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

func (c *Client) GetJobExecutionsByJobID(ctx context.Context, jobID int64) ([]*JobExecution, error) {
	var executions []*JobExecution
	if err := c.conn.WithContext(ctx).Find(&executions, "job_id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return executions, nil
}

type UpdateJobExecutionOption struct {
	ID          int64
	Status      *api.Status
	Reason      *api.Reason
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// UpdateJobExecution updates a job execution. Don't use this method directly, A helper function in dispatch.
func (c *Client) UpdateJobExecution(ctx context.Context, option *UpdateJobExecutionOption) (*JobExecution, error) {
	jobExecution, err := c.GetJobExecution(ctx, option.ID)
	if err != nil {
		return nil, err
	}
	if option.Status != nil {
		jobExecution.Status = *option.Status
	}
	if option.StartedAt != nil {
		jobExecution.StartedAt = option.StartedAt
	}
	if option.CompletedAt != nil {
		jobExecution.CompletedAt = option.CompletedAt
	}
	if option.Reason != nil {
		bys, err := json.Marshal(option.Reason)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal reason")
		}
		jobExecution.Reason = bys
	}
	return jobExecution, c.conn.WithContext(ctx).Where("id = ?", option.ID).Save(jobExecution).Error
}

func (c *Client) ResetJobExecution(ctx context.Context, jobExecutionID int64) error {
	return c.conn.WithContext(ctx).Model(&JobExecution{}).Where("id = ?", jobExecutionID).Updates(map[string]interface{}{
		"status":       api.StatusCreated,
		"reason":       "",
		"started_at":   nil,
		"completed_at": nil,
	}).Error
}
