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
	LastUpdatedAt    time.Time         `json:"last_updated_at"`
}
type JobStatus = Status

type JobExecution struct {
	ID          int64            `json:"id"`
	Status      JobStatus        `json:"status"`
	Steps       []*StepExecution `json:"steps"`
	StartedAt   *time.Time       `json:"started_at"`
	CompletedAt *time.Time       `json:"completed_at"`
}

type StepStatus = Status

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
}

type StepExecution struct {
	ID             int64      `json:"id"`
	JobExecutionID int64      `json:"job_execution_id"`
	Status         StepStatus `json:"status"`
	ExitCode       int        `json:"exit_code"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}
