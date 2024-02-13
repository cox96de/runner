package entity

import "time"

type Pipeline struct {
	ID         int64                `json:"id"`
	Executions []*PipelineExecution `json:"executions"`
	Jobs       []*Job               `json:"jobs"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

type PipelineRun struct {
	Status PipelineStatus `json:"status"`
}

type PipelineStatus = Status

type PipelineExecution struct {
	ID     int64           `json:"id"`
	Status PipelineStatus  `json:"status"`
	Steps  []*JobExecution `json:"steps"`
}
