package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cox96de/runner/api"

	"github.com/samber/lo"

	"github.com/cockroachdb/errors"
)

type Step struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID       int64     `gorm:"column:pipeline_id"`
	JobID            int64     `gorm:"column:job_id"`
	Name             string    `gorm:"column:name"`
	User             string    `gorm:"column:user"`
	Container        string    `gorm:"column:container"`
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
	Container        string
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
			Container:        option.Container,
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

// GetStepsByJobID returns steps by job id.
func (c *Client) GetStepsByJobID(ctx context.Context, jobID int64) ([]*Step, error) {
	var steps []*Step
	if err := c.conn.WithContext(ctx).Find(&steps, "job_id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return steps, nil
}

func packSteps(steps []*Step, executions []*StepExecution) ([]*api.Step, error) {
	executionsByStepID := lo.GroupBy(executions, func(item *StepExecution) int64 {
		return item.StepID
	})
	result := make([]*api.Step, 0, len(steps))
	for _, step := range steps {
		packStep, err := PackStep(step, executionsByStepID[step.ID])
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to pack step %d", step.ID)
		}
		result = append(result, packStep)
	}
	return result, nil
}

// PackStep packs a step into api.Step.
func PackStep(step *Step, executions []*StepExecution) (*api.Step, error) {
	s := &api.Step{
		ID:               step.ID,
		PipelineID:       step.PipelineID,
		JobID:            step.JobID,
		Name:             step.Name,
		WorkingDirectory: step.WorkingDirectory,
		User:             step.User,
		Container:        step.Container,
		CreatedAt:        api.ConvertTime(step.CreatedAt),
		UpdatedAt:        api.ConvertTime(step.UpdatedAt),
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
	s.Executions = lo.Map(executions, func(e *StepExecution, i int) *api.StepExecution {
		return PackStepExecution(e)
	})
	if len(s.Executions) == 0 {
		return nil, errors.Errorf("no executions found")
	}
	s.Execution = s.Executions[len(s.Executions)-1]
	return s, nil
}

// PackStepExecution packs a step execution into api.StepExecution.
func PackStepExecution(s *StepExecution) *api.StepExecution {
	return &api.StepExecution{
		ID:             s.ID,
		JobExecutionID: s.JobExecutionID,
		StepID:         s.StepID,
		Status:         s.Status,
		ExitCode:       s.ExitCode,
		StartedAt:      api.ConvertTime(s.StartedAt),
		CompletedAt:    api.ConvertTime(s.CompletedAt),
		CreatedAt:      api.ConvertTime(s.CreatedAt),
		UpdatedAt:      api.ConvertTime(s.UpdatedAt),
	}
}

type StepExecution struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	JobExecutionID int64      `gorm:"column:job_execution_id"`
	StepID         int64      `gorm:"column:step_id"`
	Status         api.Status `gorm:"column:status"`
	ExitCode       uint32     `gorm:"column:exit_code"`
	StartedAt      *time.Time `gorm:"column:started_at"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (p *StepExecution) TableName() string {
	return "step_execution"
}

type CreateStepExecutionOption struct {
	JobExecutionID int64
	StepID         int64
	Status         api.Status
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

// GetStepExecution returns step executions by id.
func (c *Client) GetStepExecution(ctx context.Context, id int64) (*StepExecution, error) {
	var step StepExecution
	if err := c.conn.WithContext(ctx).First(&step, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &step, nil
}

// GetStepExecutionsByJobExecutionID returns step executions by job execution id.
func (c *Client) GetStepExecutionsByJobExecutionID(ctx context.Context, jobExecutionID int64) ([]*StepExecution, error) {
	stepExecutions, err := c.GetStepExecutionsByJobExecutionIDs(ctx, []int64{jobExecutionID})
	if err != nil {
		return nil, err
	}
	return stepExecutions[jobExecutionID], nil
}

// GetStepExecutionsByJobExecutionIDs returns step executions by job execution id list.
func (c *Client) GetStepExecutionsByJobExecutionIDs(ctx context.Context, jobExecutionIDs []int64) (map[int64][]*StepExecution, error) {
	var steps []*StepExecution
	if err := c.conn.WithContext(ctx).Find(&steps, "job_execution_id in ?", jobExecutionIDs).Error; err != nil {
		return nil, err
	}
	return lo.GroupBy(steps, func(item *StepExecution) int64 {
		return item.JobExecutionID
	}), nil
}

type UpdateStepExecutionOption struct {
	ID          int64
	Status      *api.Status
	ExitCode    *uint32
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// UpdateStepExecution updates a step execution.
func (c *Client) UpdateStepExecution(ctx context.Context, option *UpdateStepExecutionOption) (*StepExecution, error) {
	stepExecution, err := c.GetStepExecution(ctx, option.ID)
	if err != nil {
		return nil, err
	}
	if option.Status != nil {
		stepExecution.Status = *option.Status
	}
	if option.ExitCode != nil {
		stepExecution.ExitCode = *option.ExitCode
	}
	if option.StartedAt != nil {
		stepExecution.StartedAt = option.StartedAt
	}
	if option.CompletedAt != nil {
		stepExecution.CompletedAt = option.CompletedAt
	}
	err = c.conn.WithContext(ctx).Model(&stepExecution).Where("id = ?", option.ID).Save(stepExecution).Error
	return stepExecution, err
}
