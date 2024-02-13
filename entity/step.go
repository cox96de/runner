package entity

import "time"

type StepStatus = Status

const (
	StepStatusCreated   = StatusCreated
	StepStatusRunning   = StatusRunning
	StepStatusFailed    = StatusFailed
	StepStatusSkipped   = StatusSkipped
	StepStatusSucceeded = StatusSucceeded
)

type Step struct {
	ID               int64             `json:"id"`
	PipelineID       int64             `json:"pipeline_id"`
	JobID            int64             `json:"job_id"`
	Name             string            `json:"name"`
	WorkingDirectory string            `json:"working_directory"`
	User             string            `json:"user"`
	DependsOn        []string          `json:"depends_on"`
	Commands         []string          `json:"commands"`
	EnvVar           map[string]string `json:"env_var"`
	Executions       []*StepExecution  `json:"executions"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type StepExecution struct {
	ID             int64      `json:"id"`
	JobExecutionID int64      `json:"job_execution_id"`
	Status         StepStatus `json:"status"`
	ExitCode       int        `json:"exit_code"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}
