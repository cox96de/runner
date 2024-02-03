package entity

import "time"

type Status int8

type Pipeline struct {
	ID         int64                `json:"id"`
	Jobs       []*Job               `json:"jobs"`
	Executions []*PipelineExecution `json:"executions"`
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
