package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/lo"

	"github.com/cox96de/runner/entity"

	"github.com/pkg/errors"
)

type Step struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID       int64     `gorm:"column:pipeline_id"`
	JobID            int64     `gorm:"column:job_id"`
	Name             string    `gorm:"column:uid"`
	User             string    `gorm:"column:user"`
	WorkingDirectory string    `gorm:"column:working_directory"`
	Commands         []byte    `gorm:"column:commands"`
	EnvVar           []byte    `gorm:"column:env_var"`
	DependsOn        []byte    `gorm:"column:depends_on"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *Step) TableName() string {
	return "step"
}

type CreateStepOption struct {
	PipelineID       int64
	JobID            int64
	Name             string
	User             string
	WorkingDirectory string
	EnvVar           map[string]string
	DependsOn        []string
	Commands         []string
}

// CreateSteps creates new steps.
func (c *Client) CreateSteps(ctx context.Context, options []*CreateStepOption) ([]*Step, error) {
	steps := make([]*Step, 0, len(options))
	for _, option := range options {
		step := &Step{
			PipelineID:       option.PipelineID,
			JobID:            option.JobID,
			Name:             option.Name,
			User:             option.User,
			WorkingDirectory: option.WorkingDirectory,
		}
		var err error
		step.EnvVar, err = json.Marshal(option.EnvVar)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal step.EnvVar")
		}
		step.DependsOn, err = json.Marshal(option.DependsOn)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal step.DependsOn")
		}
		step.Commands, err = json.Marshal(option.Commands)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal step.Commands")
		}
		steps = append(steps, step)
	}
	if err := c.conn.WithContext(ctx).Create(steps).Error; err != nil {
		return nil, err
	}
	return steps, nil
}

// GetStepByID returns a step by its ID.
func (c *Client) GetStepByID(ctx context.Context, id int64) (*Step, error) {
	step := &Step{}
	if err := c.conn.WithContext(ctx).First(step, id).Error; err != nil {
		return nil, err
	}
	return step, nil
}

// PackStep packs a step into entity.Step.
func PackStep(step *Step, executions []*StepExecution) (*entity.Step, error) {
	s := &entity.Step{
		ID:               step.ID,
		PipelineID:       step.PipelineID,
		JobID:            step.JobID,
		Name:             step.Name,
		User:             step.User,
		WorkingDirectory: step.WorkingDirectory,
		CreatedAt:        step.CreatedAt,
		UpdatedAt:        step.UpdatedAt,
	}
	if step.Commands != nil {
		if err := json.Unmarshal(step.Commands, &s.Commands); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal step.Commands")
		}
	}
	if step.EnvVar != nil {
		if err := json.Unmarshal(step.EnvVar, &s.EnvVar); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal step.EnvVar")
		}
	}
	if step.DependsOn != nil {
		if err := json.Unmarshal(step.DependsOn, &s.DependsOn); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal step.DependsOn")
		}
	}
	s.Executions = lo.Map(executions, func(e *StepExecution, i int) *entity.StepExecution {
		return packStepExecution(e)
	})
	return s, nil
}

func packStepExecution(s *StepExecution) *entity.StepExecution {
	return &entity.StepExecution{
		ID:             s.ID,
		JobExecutionID: s.JobExecutionID,
		Status:         s.Status,
		ExitCode:       s.ExitCode,
		StartedAt:      s.StartedAt,
		CompletedAt:    s.CompletedAt,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

type StepExecution struct {
	ID             int64             `gorm:"column:id;primaryKey;autoIncrement"`
	JobExecutionID int64             `gorm:"column:job_execution_id"`
	StepID         int64             `gorm:"column:step_id"`
	Status         entity.StepStatus `gorm:"column:status"`
	ExitCode       int               `gorm:"column:exit_code"`
	StartedAt      *time.Time        `gorm:"column:started_at"`
	CompletedAt    *time.Time        `gorm:"column:completed_at"`
	CreatedAt      time.Time         `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time         `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *StepExecution) TableName() string {
	return "step_execution"
}

type CreateStepExecutionOption struct {
	JobExecutionID int64
	StepID         int64
	Status         entity.StepStatus
}

// CreateStepExecutions creates new step executions.
func (c *Client) CreateStepExecutions(ctx context.Context, options []*CreateStepExecutionOption) ([]*StepExecution, error) {
	executions := make([]*StepExecution, 0, len(options))
	for _, option := range options {
		execution := &StepExecution{
			ID:             0,
			JobExecutionID: option.JobExecutionID,
			StepID:         option.StepID,
			Status:         option.Status,
		}
		executions = append(executions, execution)
	}
	if err := c.conn.WithContext(ctx).Create(executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}
