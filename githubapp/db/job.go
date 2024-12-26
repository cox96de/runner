package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
)

type Job struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UID                  string    `gorm:"column:uid"`
	Name                 string    `gorm:"column:name"`
	Steps                []byte    `gorm:"column:steps"`
	PipelineID           int64     `gorm:"column:pipeline_id"`
	CheckRunID           int64     `gorm:"column:check_run_id"`
	RunnerJobExecutionID int64     `gorm:"column:runner_job_execution_id"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *Job) TableName() string {
	return "job"
}

type CreateJobOption struct {
	PipelineID           int64
	Name                 string
	UID                  string
	Steps                []*Step
	CheckRunID           int64
	RunnerJobExecutionID int64
}

// Step is a snapshot of StepDSL
type Step struct {
	Name string `json:"name"`
}

func (c *Client) CreateJobs(ctx context.Context, options []*CreateJobOption) ([]*Job, error) {
	jobs := make([]*Job, 0, len(options))
	for _, opt := range options {
		job := &Job{
			PipelineID:           opt.PipelineID,
			Name:                 opt.Name,
			UID:                  opt.UID,
			CheckRunID:           opt.CheckRunID,
			RunnerJobExecutionID: opt.RunnerJobExecutionID,
		}
		bs, err := json.Marshal(opt.Steps)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal steps")
		}
		job.Steps = bs
		jobs = append(jobs, job)
	}
	if err := c.conn.WithContext(ctx).Create(jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *Client) GetJobByRunnerExecutionID(ctx context.Context, runnerExecutionID int64) (*Job, error) {
	var job Job
	if err := c.conn.WithContext(ctx).Where("runner_job_execution_id = ?", runnerExecutionID).Take(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Client) GetJobByCheckRunID(ctx context.Context, checkRunID int64) (*Job, error) {
	var job Job
	if err := c.conn.WithContext(ctx).Where("check_run_id = ?", checkRunID).Take(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

type UpdateJobOption struct {
	RunnerJobExecutionID *int64
	CheckRunID           *int64
}

func (c *Client) UpdateJob(ctx context.Context, id int64, option *UpdateJobOption) error {
	updates := map[string]interface{}{}
	if option.RunnerJobExecutionID != nil {
		updates["runner_job_execution_id"] = *option.RunnerJobExecutionID
	}
	if option.CheckRunID != nil {
		updates["check_run_id"] = *option.CheckRunID
	}
	if err := c.conn.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}
