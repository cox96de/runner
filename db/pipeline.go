package db

import (
	"context"
	"time"
)

type Pipeline struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *Pipeline) TableName() string {
	return "pipeline"
}

// CreatePipeline creates a new pipeline.
func (c *Client) CreatePipeline(ctx context.Context) (*Pipeline, error) {
	pipeline := &Pipeline{}
	if err := c.conn.WithContext(ctx).Create(pipeline).Error; err != nil {
		return nil, err
	}
	return pipeline, nil
}

type PipelineExecution struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID  int64     `gorm:"column:pipeline_id"`
	StartedAt   time.Time `gorm:"column:started_at"`
	CompletedAt time.Time `gorm:"column:completed_at"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *PipelineExecution) TableName() string {
	return "pipeline_execution"
}
