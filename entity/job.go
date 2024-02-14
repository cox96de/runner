package entity

import "time"

type Job struct {
	ID               int64             `json:"id"`
	PipelineID       int64             `json:"pipeline_id"`
	Name             string            `json:"name"`
	RunsOn           *RunsOn           `json:"runs_on"`
	WorkingDirectory string            `json:"working_directory"`
	EnvVar           map[string]string `json:"env_var"`
	DependsOn        []string          `json:"depends_on"`
	Steps            []*Step           `json:"steps"`
	Executions       []*JobExecution   `json:"executions"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
type JobStatus = Status

const (
	JobStatusCreated   = StatusCreated
	JobStatusQueued    = StatusQueued
	JobStatusRunning   = StatusRunning
	JobStatusCanceling = StatusCanceling
	JobStatusFailed    = StatusFailed
	JobStatusSkipped   = StatusSkipped
	JobStatusSucceeded = StatusSucceeded
)

type JobExecution struct {
	ID          int64            `json:"id"`
	JobID       int64            `json:"job_id"`
	Status      JobStatus        `json:"status"`
	Steps       []*StepExecution `json:"steps"`
	StartedAt   *time.Time       `json:"started_at"`
	CompletedAt *time.Time       `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
