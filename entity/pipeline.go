package entity

import "time"

type Status int8

type Pipeline struct {
	ID            int64                `json:"id"`
	Executions    []*PipelineExecution `json:"executions"`
	CreatedAt     time.Time            `json:"created_at"`
	LastUpdatedAt time.Time            `json:"last_updated_at"`
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
