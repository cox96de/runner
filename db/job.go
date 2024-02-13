package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cox96de/runner/entity"
	"github.com/pkg/errors"
)

type Job struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID       int64     `gorm:"column:pipeline_id"`
	Name             string    `gorm:"column:uid"`
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
	RunsOn           *entity.RunsOn
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

// PackJob packs a job into entity.Job.
func PackJob(j *Job) (*entity.Job, error) {
	runsOn := &entity.RunsOn{}
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
	return &entity.Job{
		ID:               j.ID,
		PipelineID:       j.PipelineID,
		Name:             j.Name,
		RunsOn:           runsOn,
		WorkingDirectory: j.WorkingDirectory,
		EnvVar:           envVar,
		DependsOn:        dependsOn,
		CreatedAt:        j.CreatedAt,
		UpdatedAt:        j.UpdatedAt,
	}, nil
}

type JobExecution struct {
	ID    int64 `gorm:"column:id;primaryKey;autoIncrement"`
	JobID int64 `gorm:"column:job_id"`
	// TODO: Status type
	Status      entity.JobStatus `gorm:"column:status"`
	StartedAt   time.Time        `gorm:"column:started_at"`
	CompletedAt time.Time        `gorm:"column:completed_at"`
	CreatedAt   time.Time        `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time        `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *JobExecution) TableName() string {
	return "job_execution"
}
