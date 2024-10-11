package db

import (
	"context"
	"time"
)

type Pipeline struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	AppInstallID int64     `gorm:"column:app_install_id"`
	RepoOwner    string    `gorm:"column:repo_owner"`
	RepoName     string    `gorm:"column:repo_name"`
	HeadSHA      string    `gorm:"column:head_sha"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *Pipeline) TableName() string {
	return "pipeline"
}

type CreatePipelineOption struct {
	AppInstallID int64
	RepoOwner    string
	RepoName     string
	HeadSHA      string
}

// CreatePipeline creates a new pipeline.
func (c *Client) CreatePipeline(ctx context.Context, opt *CreatePipelineOption) (*Pipeline, error) {
	pipeline := &Pipeline{
		AppInstallID: opt.AppInstallID,
		RepoOwner:    opt.RepoOwner,
		RepoName:     opt.RepoName,
		HeadSHA:      opt.HeadSHA,
	}
	if err := c.conn.WithContext(ctx).Create(pipeline).Error; err != nil {
		return nil, err
	}
	return pipeline, nil
}

// GetPipelineByID gets a pipeline by ID.
func (c *Client) GetPipelineByID(ctx context.Context, id int64) (*Pipeline, error) {
	pipeline := &Pipeline{}
	if err := c.conn.WithContext(ctx).Where("id = ?", id).First(pipeline).Error; err != nil {
		return nil, err
	}
	return pipeline, nil
}
