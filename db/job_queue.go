package db

import (
	"context"
	"time"

	"github.com/cox96de/runner/api"
)

type JobQueue struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Status         api.Status `gorm:"column:status"`
	JobExecutionID int64      `gorm:"column:job_execution_id"`
	Label          string     `gorm:"column:label"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *JobQueue) TableName() string {
	return "job_queue"
}

type CreateJobQueueOption struct {
	JobExecutionID int64
	Label          string
	Status         api.Status
}

// CreateJobQueues creates new job queue entity.
func (c *Client) CreateJobQueues(ctx context.Context, options []*CreateJobQueueOption) ([]*JobQueue, error) {
	jobQueues := make([]*JobQueue, 0, len(options))
	for _, opt := range options {
		job := &JobQueue{
			JobExecutionID: opt.JobExecutionID,
			Label:          opt.Label,
			Status:         opt.Status,
		}
		jobQueues = append(jobQueues, job)
	}
	if err := c.conn.WithContext(ctx).Create(jobQueues).Error; err != nil {
		return nil, err
	}
	return jobQueues, nil
}

func (c *Client) GetJobQueue(ctx context.Context, jobExecutionID int64) (*JobQueue, error) {
	jobQueue := &JobQueue{}
	if err := c.conn.WithContext(ctx).First(jobQueue, "job_execution_id = ?", jobExecutionID).Error; err != nil {
		return nil, err
	}
	return jobQueue, nil
}

func (c *Client) UpdateJobQueueStatus(ctx context.Context, jobExecutionID int64, status api.Status) error {
	updateField := map[string]interface{}{
		"status": status,
	}
	return c.conn.WithContext(ctx).Model(&JobQueue{}).Where("job_execution_id = ?", jobExecutionID).Updates(updateField).Error
}

func (c *Client) DeleteJobQueueByJobExecutionID(ctx context.Context, jobExecutionID int64) error {
	if err := c.conn.WithContext(ctx).
		Where("job_execution_id = ?", jobExecutionID).Delete(&JobQueue{}).Error; err != nil {
		return err
	}
	return nil
}

func (c *Client) GetQueuedJobExecutionIDs(ctx context.Context, label string, limit int) ([]*JobQueue, error) {
	executions := make([]*JobQueue, 0, limit)
	if err := c.conn.WithContext(ctx).Where("label = ? and status = ? order by id", label, api.StatusQueued).
		Limit(limit).Find(&executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}
